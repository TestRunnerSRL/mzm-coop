package main

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"sync"
	"testing"
)

type FakeRoom struct {
	kicked []string
	err    error
}

func (r *FakeRoom) Kick(name string) error {
	r.kicked = append(r.kicked, name)
	return r.err
}

type FakePlayerListSubscriber struct {
	subs    []chan string
	unsubs  []chan string
	pending sync.WaitGroup
}

func (pl *FakePlayerListSubscriber) Subscribe(c chan string) {
	pl.subs = append(pl.subs, c)
	c <- "Initial"
	pl.pending.Add(1)
}
func (pl *FakePlayerListSubscriber) Unsubscribe(c chan string) {
	pl.unsubs = append(pl.unsubs, c)
	pl.pending.Done()
}

const PORT = 12345

func TestRoot(t *testing.T) {
	s := NewAdminServer(PORT, &FakeRoom{}, &FakePlayerListSubscriber{})
	defer s.Close()
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/", PORT))
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("get failed: %s", resp.Status)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if len(body) == 0 {
		t.Errorf("got %v, want ''", string(body))
	}
}

func TestKickSuccess(t *testing.T) {
	r := FakeRoom{}
	s := NewAdminServer(PORT, &r, &FakePlayerListSubscriber{})
	defer s.Close()
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/kick/foo", PORT))
	if err != nil {
		t.Fatalf("kick failed: %v", err)
	}
	defer resp.Body.Close()
	want := []string{"foo"}
	if !reflect.DeepEqual(r.kicked, want) {
		t.Errorf("got %v, want %v", r.kicked, want)
	}
}

func TestKickFailure(t *testing.T) {
	r := FakeRoom{err: errors.New("")}
	s := NewAdminServer(PORT, &r, &FakePlayerListSubscriber{})
	defer s.Close()
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/kick/foo", PORT))
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("got %v, want %v", resp.StatusCode, http.StatusNotFound)
	}
}

func TestPlayerlistSubscribing(t *testing.T) {
	pl := FakePlayerListSubscriber{}
	s := NewAdminServer(PORT, &FakeRoom{}, &pl)
	defer s.Close()
	req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:%d/playerlist", PORT), nil)
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}

	headers := []struct {
		header string
		want   string
	}{
		{"Content-Type", "text/event-stream"},
		{"Cache-Control", "no-cache"},
		{"Connection", "keep-alive"},
		{"Access-Control-Allow-Origin", "*"},
	}
	for _, tt := range headers {
		if h := resp.Header.Get(tt.header); h != tt.want {
			t.Errorf("got %s %s, want %s", tt.header, h, tt.want)
		}
	}

	scanner := bufio.NewScanner(resp.Body)
	want := "data: Initial"
	if scanner.Scan(); scanner.Text() != want {
		t.Errorf("got %s, want %s", scanner.Text(), want)
	}
	if scanner.Scan(); scanner.Text() != "" {
		t.Errorf("got %s, want ''", scanner.Text())
	}

	if len(pl.subs) != 1 {
		t.Fatalf("channel not subscribed")
	}
	pl.subs[0] <- "Foo"
	want = "data: Foo"
	if scanner.Scan(); scanner.Text() != want {
		t.Errorf("got %s, want %s", scanner.Text(), want)
	}
	if scanner.Scan(); scanner.Text() != "" {
		t.Errorf("got %s, want ''", scanner.Text())
	}

	resp.Body.Close()
	pl.pending.Wait()
	if !reflect.DeepEqual(pl.unsubs, pl.subs) {
		t.Errorf("got %v, want %v", pl.unsubs, pl.subs)
	}
}
