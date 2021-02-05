package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"
)

type FakeConn struct {
	input  io.Reader
	output io.Writer
	closed bool
}

func (f *FakeConn) Read(buf []byte) (int, error) {
	return f.input.Read(buf)
}

func (f *FakeConn) Write(buf []byte) (int, error) {
	return f.output.Write(buf)
}

func (f *FakeConn) Close() error {
	f.closed = true
	if r, ok := f.input.(io.ReadCloser); ok {
		r.Close()
	}
	if w, ok := f.output.(io.WriteCloser); ok {
		w.Close()
	}
	return nil
}

type FakeSyncConfig struct {
	validateErr   error
	clientConfigs []*Message
	releaseCalls  int
}

func (sc *FakeSyncConfig) ConfigMessagePayload(id PlayerID) string {
	return fmt.Sprintf("configPayload%d", id)
}
func (sc *FakeSyncConfig) Itemlist() string         { return "itemPayload" }
func (sc *FakeSyncConfig) ReleasePlayerID(PlayerID) { sc.releaseCalls++ }
func (sc *FakeSyncConfig) ValidateClientConfig(msg *Message) (PlayerID, error) {
	sc.clientConfigs = append(sc.clientConfigs, msg)
	return PlayerID(len(sc.clientConfigs)), sc.validateErr
}

type FakePlayerList struct {
	statuses         []*Message
	validateErr      error
	playerNums       []*Message
	releaseCalls     int
	subscribers      []chan string
	unsubscribeCalls int
}

func (pl *FakePlayerList) UpdateStatus(msg *Message)  { pl.statuses = append(pl.statuses, msg) }
func (pl *FakePlayerList) ReleasePlayerNum(PlayerNum) { pl.releaseCalls++ }
func (pl *FakePlayerList) ValidatePlayerNumber(msg *Message) (PlayerNum, error) {
	pl.playerNums = append(pl.playerNums, msg)
	return PlayerNum(len(pl.playerNums)), pl.validateErr
}
func (pl *FakePlayerList) Subscribe(c chan string) {
	c <- "playerListPayload"
	pl.subscribers = append(pl.subscribers, c)
}
func (pl *FakePlayerList) Unsubscribe(chan string) { pl.unsubscribeCalls++ }

func TestMessagePropagation(t *testing.T) {
	sc := &FakeSyncConfig{}
	pl := &FakePlayerList{}
	room := NewRoom(sc, pl)
	wg := sync.WaitGroup{}
	r1, writer1 := io.Pipe()
	reader1, w1 := io.Pipe()
	conn1 := FakeConn{input: r1, output: w1}
	go func() {
		room.HandleConnection(&conn1)
		wg.Done()
	}()

	// First the server expects our config.
	if _, err := fmt.Fprintf(writer1, "cP1,syncHash\n"); err != nil {
		t.Errorf("conn1: failed to write config: %v", err)
	}

	// Then the server sends its config.
	configMessage := "cServer,configPayload1"
	scanner1 := bufio.NewScanner(reader1)
	if scanner1.Scan(); scanner1.Text() != configMessage {
		t.Errorf("conn1: got config %s, want %s", scanner1.Text(), configMessage)
	}

	// Then the server sends its player number.
	playerNumberMessage := "nServer,Server,0"
	if scanner1.Scan(); scanner1.Text() != playerNumberMessage {
		t.Errorf("conn1: got player number %s, want %s", scanner1.Text(), playerNumberMessage)
	}

	// Then the server expects our player number.
	if _, err := fmt.Fprintf(writer1, "nP1,P1,1\n"); err != nil {
		t.Errorf("conn1: failed to write player number: %v", err)
	}

	// Then the server sends the itemlist.
	itemList := "rServer,itemPayload"
	if scanner1.Scan(); scanner1.Text() != itemList {
		t.Errorf("conn1: got itemlist %s, want %s", scanner1.Text(), itemList)
	}

	// Then the server sends the playerlist.
	playerList := "lServer,playerListPayload"
	if scanner1.Scan(); scanner1.Text() != playerList {
		t.Errorf("conn1: got playerlist %s, want %s", scanner1.Text(), playerList)
	}

	r2, writer2 := io.Pipe()
	reader2, w2 := io.Pipe()
	conn2 := FakeConn{input: r2, output: w2}
	go func() {
		room.HandleConnection(&conn2)
		wg.Done()
	}()

	scanner2 := bufio.NewScanner(reader2)
	if _, err := fmt.Fprintf(writer2, "cP2,syncHash\n"); err != nil {
		t.Errorf("conn2: failed to write config: %v", err)
	}
	configMessage = "cServer,configPayload2"
	if scanner2.Scan(); scanner2.Text() != configMessage {
		t.Errorf("conn2: got config %s, want %s", scanner2.Text(), configMessage)
	}
	if scanner2.Scan(); scanner2.Text() != playerNumberMessage {
		t.Errorf("conn2: got player number %s, want %s", scanner2.Text(), playerNumberMessage)
	}
	if _, err := fmt.Fprintf(writer2, "nP2,P2,2\n"); err != nil {
		t.Errorf("conn2: failed to write player number: %v", err)
	}
	if scanner2.Scan(); scanner2.Text() != itemList {
		t.Errorf("conn2: got itemlist %s, want %s", scanner2.Text(), itemList)
	}
	if scanner2.Scan(); scanner2.Text() != playerList {
		t.Errorf("conn2: got playerlist %s, want %s", scanner2.Text(), playerList)
	}

	// Write a message to P1. It should be forwarded to P2.
	want := "mP1,payload"
	if _, err := fmt.Fprintf(writer1, "%s\n", want); err != nil {
		t.Errorf("conn1: failed to message: %v", err)
	}
	if scanner2.Scan(); scanner2.Text() != want {
		t.Errorf("conn2: got message %s, want %s", scanner2.Text(), want)
	}

	// Close the inputs. HandleConnection should close the outputs before returning.
	wg.Add(2)
	writer1.Close()
	writer2.Close()
	wg.Wait()
	if !conn1.closed || !conn2.closed {
		t.Errorf("connections were not closed")
	}

	// Verify the configs sent to the SyncConfig.
	var wantConfigs = []*Message{
		&Message{ConfigMessageType, "P1", "syncHash"},
		&Message{ConfigMessageType, "P2", "syncHash"},
	}
	if !reflect.DeepEqual(sc.clientConfigs, wantConfigs) {
		t.Errorf("unexpected config messages: got %v, want %v", sc.clientConfigs, wantConfigs)
	}

	// Verify the messages sent to the PlayerList.
	var wantPlayerNums = []*Message{
		&Message{PlayerNumberMessageType, "P1", "P1,1"},
		&Message{PlayerNumberMessageType, "P2", "P2,2"},
	}
	if !reflect.DeepEqual(pl.playerNums, wantPlayerNums) {
		t.Errorf("unexpected player number messages: got %v, want %v", pl.playerNums, wantPlayerNums)
	}
}

func TestStatusMessage(t *testing.T) {
	pl := FakePlayerList{}
	room := NewRoom(&FakeSyncConfig{}, &pl)
	input := "cP,syncHash\nnP,P,\nsP,P,Ready\n"
	conn := FakeConn{input: bytes.NewBufferString(input), output: new(bytes.Buffer)}
	room.HandleConnection(&conn)
	if len(pl.statuses) != 1 {
		t.Fatalf("Unexpected number of status messages: got %d, want 1", len(pl.statuses))
	}
	want := &Message{
		MessageType:  PlayerStatusMessageType,
		FromUserName: "P",
		Payload:      "P,Ready",
	}
	if !reflect.DeepEqual(pl.statuses[0], want) {
		t.Errorf("got %v, want %v", pl.statuses[0], want)
	}
}

func TestKick(t *testing.T) {
	room := NewRoom(&FakeSyncConfig{}, &FakePlayerList{})
	r1, writer1 := io.Pipe()
	reader1, w1 := io.Pipe()
	conn1 := FakeConn{input: r1, output: w1}
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		room.HandleConnection(&conn1)
		wg.Done()
	}()
	scanner1 := bufio.NewScanner(reader1)
	if _, err := fmt.Fprintf(writer1, "cP1,syncHash\n"); err != nil {
		t.Errorf("conn1: failed to write config: %v", err)
	}
	if scanner1.Scan(); !strings.HasPrefix(scanner1.Text(), "c") {
		t.Fatalf("conn1: got %s, wanted config message", scanner1.Text())
	}
	if scanner1.Scan(); !strings.HasPrefix(scanner1.Text(), "n") {
		t.Fatalf("conn1: got %s, wanted player number", scanner1.Text())
	}
	if _, err := fmt.Fprintf(writer1, "nP1,P1,\n"); err != nil {
		t.Errorf("conn1: failed to write player number: %v", err)
	}
	if scanner1.Scan(); !strings.HasPrefix(scanner1.Text(), "r") {
		t.Fatalf("conn1: got %s, wanted itemlist", scanner1.Text())
	}
	if scanner1.Scan(); !strings.HasPrefix(scanner1.Text(), "l") {
		t.Fatalf("conn1: got %s, wanted playerlist", scanner1.Text())
	}

	r2, writer2 := io.Pipe()
	reader2, w2 := io.Pipe()
	conn2 := FakeConn{input: r2, output: w2}
	go room.HandleConnection(&conn2)
	scanner2 := bufio.NewScanner(reader2)
	if _, err := fmt.Fprintf(writer2, "cP2,syncHash\n"); err != nil {
		t.Fatalf("conn2: failed to write config: %v", err)
	}
	if scanner2.Scan(); !strings.HasPrefix(scanner2.Text(), "c") {
		t.Fatalf("conn2: got %s, wanted config message", scanner2.Text())
	}
	if scanner2.Scan(); !strings.HasPrefix(scanner2.Text(), "n") {
		t.Fatalf("conn2: got %s, wanted player number", scanner2.Text())
	}
	if _, err := fmt.Fprintf(writer2, "nP2,P2,\n"); err != nil {
		t.Fatalf("conn2: failed to write player number: %v", err)
	}
	if scanner2.Scan(); !strings.HasPrefix(scanner2.Text(), "r") {
		t.Fatalf("conn2: got %s, wanted itemlist", scanner2.Text())
	}
	if scanner2.Scan(); !strings.HasPrefix(scanner2.Text(), "l") {
		t.Fatalf("conn2: got %s, wanted playerlist", scanner2.Text())
	}
	if scanner1.Scan(); !strings.HasPrefix(scanner1.Text(), "r") {
		t.Fatalf("conn1: got %s, wanted itemlist", scanner1.Text())
	}

	// Kicking an invalid player should fail.
	if err := room.Kick("P3"); !errors.Is(err, ErrUnknownUser) {
		t.Errorf("got %v, want %v", err, ErrUnknownUser)
	}

	// Kicking a valid player should cause that player to receive a kick
	// message and other players to receive a quit message with 'q:was_kicked'.
	if err := room.Kick("P1"); err != nil {
		t.Fatalf("kick failed: %v", err)
	}
	want := "kServer,"
	if scanner1.Scan(); scanner1.Text() != want {
		t.Fatalf("conn1: got %s, want %s", scanner1.Text(), want)
	}
	want = "qP1,q:was_kicked"
	if scanner2.Scan(); scanner2.Text() != want {
		t.Fatalf("conn2: got %s, want %s", scanner2.Text(), want)
	}

	// P1's connection should also be closed.
	wg.Wait()
}

func TestPlayerListSubscribe(t *testing.T) {
	pl := FakePlayerList{}
	room := NewRoom(&FakeSyncConfig{}, &pl)
	r, writer := io.Pipe()
	reader, w := io.Pipe()
	conn := FakeConn{input: r, output: w}
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		room.HandleConnection(&conn)
		wg.Done()
	}()

	scanner := bufio.NewScanner(reader)
	if _, err := fmt.Fprintf(writer, "cP1,syncHash\n"); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}
	if scanner.Scan(); !strings.HasPrefix(scanner.Text(), "c") {
		t.Fatalf("got %s, wanted config message", scanner.Text())
	}
	if scanner.Scan(); !strings.HasPrefix(scanner.Text(), "n") {
		t.Fatalf("got %s, wanted player number", scanner.Text())
	}
	if _, err := fmt.Fprintf(writer, "nP1,P1,\n"); err != nil {
		t.Fatalf("failed to write player number: %v", err)
	}
	if scanner.Scan(); !strings.HasPrefix(scanner.Text(), "r") {
		t.Fatalf("got %s, wanted itemlist", scanner.Text())
	}
	if scanner.Scan(); !strings.HasPrefix(scanner.Text(), "l") {
		t.Fatalf("got %s, wanted playerlist", scanner.Text())
	}

	// Writing to the channel should cause the playerlist to be sent.
	if len(pl.subscribers) != 1 {
		t.Fatalf("Unexpected number of Subscribe calls: %d", len(pl.subscribers))
	}
	pl.subscribers[0] <- "NewPlayerList"
	want := "lServer,NewPlayerList"
	if scanner.Scan(); scanner.Text() != want {
		t.Fatalf("got %s, want %s", scanner.Text(), want)
	}

	// Close the connection; the channel should be unsubscribed.
	writer.Close()
	wg.Wait()
	if pl.unsubscribeCalls != 1 {
		t.Errorf("got %d Unsubscribe() call, want 1", pl.unsubscribeCalls)
	}
}

func TestClose(t *testing.T) {
	room := NewRoom(&FakeSyncConfig{}, &FakePlayerList{})
	reader, writer := io.Pipe()
	var buf bytes.Buffer
	conn := FakeConn{input: reader, output: &buf}
	go room.HandleConnection(&conn)

	if _, err := fmt.Fprintf(writer, "cP1,syncHash\n"); err != nil {
		t.Errorf("failed to write config: %v", err)
	}
	if _, err := fmt.Fprintf(writer, "nP1,P1,\n"); err != nil {
		t.Errorf("failed to write player number: %v", err)
	}
	// Write another message to ensure we're beyond the initial handshake.
	if _, err := fmt.Fprintf(writer, "pP1,\n"); err != nil {
		t.Errorf("failed to write ping: %v", err)
	}

	room.Close()
	if !conn.closed {
		t.Errorf("connection was not closed")
	}
	want := "cServer,configPayload1\n" +
		"nServer,Server,0\n" +
		"rServer,itemPayload\n" +
		"lServer,playerListPayload\n" +
		"qServer,q:\n"
	if buf.String() != want {
		t.Errorf("got %s, want %s", buf.String(), want)
	}
}

func TestSyncFailures(t *testing.T) {
	var tests = []struct {
		input          string
		syncError      error
		playerNumError error
		want           error
	}{
		{"", nil, nil, ErrConnectionClosed},
		{"x,\n", nil, nil, ErrUnknownMessageType},
		{"c,syncHash\n", ErrSyncBadHash, nil, ErrSyncBadHash},
		{"c,syncHash\n", nil, nil, ErrConnectionClosed},
		{"c,syncHash\nx,\n", nil, nil, ErrUnknownMessageType},
		{"c,syncHash\nn,P,\n", nil, ErrPlayerNumPayload, ErrPlayerNumPayload},
	}

	defer log.SetOutput(os.Stderr)
	for _, tt := range tests {
		var buf bytes.Buffer
		log.SetOutput(&buf)
		room := NewRoom(&FakeSyncConfig{validateErr: tt.syncError}, &FakePlayerList{validateErr: tt.playerNumError})
		conn := FakeConn{input: bytes.NewBufferString(tt.input), output: new(bytes.Buffer)}
		room.HandleConnection(&conn)
		if !conn.closed {
			t.Errorf("connection wasn't closed for %s", tt.input)
		}
		if !strings.Contains(buf.String(), tt.want.Error()) {
			t.Errorf("expected %s in logs, got %s", tt.want.Error(), buf.String())
		}
	}
}

func TestDuplicateUserName(t *testing.T) {
	room := NewRoom(&FakeSyncConfig{}, &FakePlayerList{})
	done := make(chan bool, 1)

	r1, writer1 := io.Pipe()
	conn1 := FakeConn{input: r1, output: new(bytes.Buffer)}
	go room.HandleConnection(&conn1)
	if _, err := fmt.Fprintf(writer1, "cP,syncHash\n"); err != nil {
		t.Errorf("conn1: failed to write config: %v", err)
	}
	if _, err := fmt.Fprintf(writer1, "nP,P\n"); err != nil {
		t.Errorf("conn1: failed to write player number: %v", err)
	}
	// Write another message to ensure we're beyond the initial handshake.
	if _, err := fmt.Fprintf(writer1, "pP1,\n"); err != nil {
		t.Errorf("failed to write ping: %v", err)
	}

	r2, writer2 := io.Pipe()
	conn2 := FakeConn{input: r2, output: new(bytes.Buffer)}
	go func() {
		room.HandleConnection(&conn2)
		done <- true
	}()
	var buf bytes.Buffer
	log.SetOutput(&buf)
	if _, err := fmt.Fprintf(writer2, "cP,syncHash\n"); err != nil {
		t.Errorf("conn2: failed to write config: %v", err)
	}
	if _, err := fmt.Fprintf(writer2, "nP,P\n"); err != nil {
		t.Errorf("conn2: failed to write player number: %v", err)
	}
	<-done
	log.SetOutput(os.Stderr)

	// P2s connection should have closed with an error.
	if !conn2.closed {
		t.Errorf("conn2 was not closed")
	}
	if !strings.Contains(buf.String(), ErrSyncUserNameInUse.Error()) {
		t.Errorf("expected %s in logs, got %s", ErrSyncUserNameInUse.Error(), buf.String())
	}
	if conn1.closed {
		t.Errorf("conn was not closed")
	}
}

// Note: pings currently aren't tested due to time.Ticker being difficult to
// fake in tests.
