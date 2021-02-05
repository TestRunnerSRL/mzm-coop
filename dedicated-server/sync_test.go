package main

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestConfigMessagePayload(t *testing.T) {
	sc := NewSyncConfig("syncHash", "ramConfig", 1)
	want := "syncHash,3,ramConfig"
	if payload := sc.ConfigMessagePayload(PlayerID(3)); payload != want {
		t.Errorf("got %s, want %s", payload, want)
	}
}

func TestLazyConfigMessagePayload(t *testing.T) {
	sc := NewSyncConfig("", "ramConfig", 1)
	want := ",3,ramConfig"
	if payload := sc.ConfigMessagePayload(PlayerID(3)); payload != want {
		t.Errorf("got %s, want %s", payload, want)
	}

	// A player with an unknown synchash should be rejected. The synchash
	// should not change.
	msg := &Message{ConfigMessageType, "", "syncHash"}
	if _, err := sc.ValidateClientConfig(msg); !errors.Is(err, ErrSyncBadHash) {
		t.Errorf("config should have been rejected: got %v, want %v", err, ErrSyncBadHash)
	}
	if payload := sc.ConfigMessagePayload(PlayerID(3)); payload != want {
		t.Errorf("got %s, want %s", payload, want)
	}

	// A player with a known synchash should be accepted and the server'same
	// synchash should be updated.
	for syncHash := range SyncHashes {
		msg.Payload = syncHash
		break
	}
	if _, err := sc.ValidateClientConfig(msg); err != nil {
		t.Errorf("failed to validate config message: %v", err)
	}
	want = msg.Payload + ",3,ramConfig"
	if payload := sc.ConfigMessagePayload(PlayerID(3)); payload != want {
		t.Errorf("got %s, want %s", payload, want)
	}
}

func TestItemlistWithoutPlayers(t *testing.T) {
	sc := NewSyncConfig("syncHash", "ramConfig", 2)
	want := "i:0:1,i:1:1"
	if itemlist := sc.Itemlist(); itemlist != want {
		t.Errorf("got %s, want %s", itemlist, want)
	}
}

func TestItemlist(t *testing.T) {
	const itemcount = 9
	sc := NewSyncConfig("syncHash", "ramConfig", itemcount)
	msg := &Message{ConfigMessageType, "", "syncHash"}
	ids := make([]PlayerID, 4)
	for i := range ids {
		var err error
		if ids[i], err = sc.ValidateClientConfig(msg); err != nil {
			t.Errorf("failed to validate config message: %v", err)
		}
	}
	// Remove player 2 so that there's a gap in player ids.
	sc.ReleasePlayerID(ids[2])

	// Since the itemlist is shuffled, we can't test it against a specific
	// string. Instead, verify that each player receives the right number of
	// items and that all items are assigned.
	parts := strings.Split(sc.Itemlist(), ",")
	if len(parts) != itemcount {
		t.Errorf("wrong item count: got %d, want %d", len(parts), itemcount)
	}
	perItemCounts := make(map[int]int)
	perPlayerCounts := make(map[PlayerID]int)
	for _, part := range parts {
		var item, id int
		if n, err := fmt.Sscanf(part, "i:%d:%d", &item, &id); err != nil || n != 2 {
			t.Errorf("parsed %d values from %s with error %v", n, part, err)
		}
		perItemCounts[item]++
		perPlayerCounts[PlayerID(id)]++
	}

	// Each item [0,itemcount) should appear exactly once.
	wantItemCounts := make(map[int]int)
	for i := 0; i < itemcount; i++ {
		wantItemCounts[i] = 1
	}
	if !reflect.DeepEqual(perItemCounts, wantItemCounts) {
		t.Errorf("unexpected item counts: got %v, want %v", perItemCounts, wantItemCounts)
	}
	// Each player should have the same number of items.
	wantPlayerCounts := map[PlayerID]int{
		ids[0]: itemcount / (len(ids) - 1),
		ids[1]: itemcount / (len(ids) - 1),
		ids[3]: itemcount / (len(ids) - 1),
	}
	if !reflect.DeepEqual(perPlayerCounts, wantPlayerCounts) {
		t.Errorf("unexpected player counts: got %v, want %v", perPlayerCounts, wantPlayerCounts)
	}
}

func TestValidateClientConfig(t *testing.T) {
	var tests = []struct {
		msg  *Message
		want error
	}{
		{&Message{ConfigMessageType, "", "syncHash"}, nil},
		{&Message{MemoryMessageType, "", "syncHash"}, ErrSyncWrongMessageType},
		{&Message{ConfigMessageType, "", "badHash"}, ErrSyncBadHash},
	}

	for _, tt := range tests {
		sc := NewSyncConfig("syncHash", "ramConfig", 1)
		if _, err := sc.ValidateClientConfig(tt.msg); !errors.Is(err, tt.want) {
			t.Errorf("got %v, want %v", err, tt.want)
		}
	}
}

func TestPlayerIDs(t *testing.T) {
	sc := NewSyncConfig("syncHash", "ramConfig", 1)
	msg := &Message{ConfigMessageType, "", "syncHash"}
	ids := make([]PlayerID, 5)
	for i := range ids {
		var err error
		if ids[i], err = sc.ValidateClientConfig(msg); err != nil {
			t.Errorf("failed to validate config message: %v", err)
		}
	}
	want := []PlayerID{2, 3, 4, 5, 6}
	if !reflect.DeepEqual(ids, want) {
		t.Errorf("got %v, want %v", ids, want)
	}

	// Release ids 3 and 5; the next validation call should return these
	// values (in order) before returning 7.
	sc.ReleasePlayerID(3)
	sc.ReleasePlayerID(5)
	ids = make([]PlayerID, 3)
	for i := range ids {
		var err error
		if ids[i], err = sc.ValidateClientConfig(msg); err != nil {
			t.Errorf("failed to validate config message: %v", err)
		}
	}
	want = []PlayerID{3, 5, 7}
	if !reflect.DeepEqual(ids, want) {
		t.Errorf("got %v, want %v", ids, want)
	}
}
