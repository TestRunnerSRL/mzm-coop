package main

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
)

var (
	// ErrPlayerNumWrongMessageType is returned when ValidatePlayerNumber()
	// is called with something other than a player number message.
	ErrPlayerNumWrongMessageType = errors.New("wrong message type")
	// ErrPlayerNumPayload is returned when ValidatePlayerNumber() is called
	// with an invalid player number message payload.
	ErrPlayerNumPayload = errors.New("invalid message payload")
	// ErrPlayerNumInUse is returned when a player attempts to use a PlayerNum
	// that's already in use by a different player.
	ErrPlayerNumInUse = errors.New("player number in use")
)

const defaultStatus = "Unready"

// PlayerList keeps track of which players are connected.
type PlayerList struct {
	mux         sync.Mutex
	players     map[PlayerNum]string // guarded by mux
	statuses    map[string]string    // guarded by mux
	subscribers []chan string        // guarded by mux
}

// PlayerNum is a unique number for each player.
type PlayerNum int

// NewPlayerList creates a new Playerlist object.
func NewPlayerList() *PlayerList {
	return &PlayerList{
		players:  make(map[PlayerNum]string),
		statuses: make(map[string]string),
	}
}

// Returns a payload string encoding the currently connected players. The
// server is never included. Requires pl.mux to be held.
func (pl *PlayerList) playerlist() string {
	var sb strings.Builder
	for num, name := range pl.players {
		sb.WriteString(fmt.Sprintf("l:%s:num:%d,l:%s:status:%s,", name, num, name, pl.statuses[name]))
	}
	if sb.Len() == 0 {
		return ""
	}
	return sb.String()[0 : sb.Len()-1]
}

// UpdateStatus uses a player status messages and updates the corresponding
// player's status. All subscribers will be notified.
func (pl *PlayerList) UpdateStatus(msg *Message) {
	if msg.MessageType != PlayerStatusMessageType {
		log.Printf("UpdateStatus called with invalid message type: %v", msg)
		return
	}

	parts := strings.SplitN(msg.Payload, ",", 2)
	if parts[0] != msg.FromUserName || len(parts) != 2 {
		log.Printf("Ignoring invalid player status message: %v", msg)
		return
	}

	// Ignore status messages from unknown players.
	pl.mux.Lock()
	if _, ok := pl.statuses[msg.FromUserName]; ok {
		pl.statuses[msg.FromUserName] = parts[1]
		pl.notifySubscribers()
	}
	pl.mux.Unlock()
}

// ValidatePlayerNumber validates the player number message sent by clients,
// ensuring that the self-reported number is unique.  Since this changes the
// list of connected players, all subscribers will be notified.
func (pl *PlayerList) ValidatePlayerNumber(msg *Message) (PlayerNum, error) {
	if msg.MessageType != PlayerNumberMessageType {
		return 0, ErrPlayerNumWrongMessageType
	}

	// Determine the player number. If it's empty, the server is supposed to
	// pick the next available value.
	parts := strings.SplitN(msg.Payload, ",", 2)
	if parts[0] != msg.FromUserName {
		return 0, ErrPlayerNumPayload
	}
	if len(parts) == 1 || parts[1] == "" {
		pl.mux.Lock()
		for num := PlayerNum(1); ; num++ {
			if _, ok := pl.players[num]; !ok {
				pl.players[num] = msg.FromUserName
				pl.statuses[msg.FromUserName] = defaultStatus
				pl.notifySubscribers()
				pl.mux.Unlock()
				return num, nil
			}
		}
	} else {
		num, err := strconv.Atoi(parts[1])
		if err != nil {
			return 0, fmt.Errorf("bad player number: %w", err)
		}
		pl.mux.Lock()
		defer pl.mux.Unlock()
		if name, ok := pl.players[PlayerNum(num)]; ok {
			return 0, fmt.Errorf("%w by %s", ErrPlayerNumInUse, name)
		}
		pl.players[PlayerNum(num)] = msg.FromUserName
		pl.statuses[msg.FromUserName] = defaultStatus
		pl.notifySubscribers()
		return PlayerNum(num), nil
	}
}

// ReleasePlayerNum returns the PlayerNum to the list of available player
// numbers. Since this changes the list of connected players, all subscribers
// will be notified.
func (pl *PlayerList) ReleasePlayerNum(num PlayerNum) {
	pl.mux.Lock()
	if name, ok := pl.players[num]; ok {
		delete(pl.statuses, name)
		delete(pl.players, num)
		pl.notifySubscribers()
	}
	pl.mux.Unlock()
}

// Subscribe causes the channel to receive updates to the playerlist until
// Unsubscribe() is called. The current playerlist is sent immediately.
func (pl *PlayerList) Subscribe(c chan string) {
	pl.mux.Lock()
	pl.subscribers = append(pl.subscribers, c)
	c <- pl.playerlist()
	pl.mux.Unlock()
}

// Unsubscribe stops sending updates to the channel.
func (pl *PlayerList) Unsubscribe(c chan string) {
	pl.mux.Lock()
	for i := 0; i < len(pl.subscribers); i++ {
		if pl.subscribers[i] == c {
			// If we find the channel, replace it with the last element in the
			// slice and resize.
			pl.subscribers[i] = pl.subscribers[len(pl.subscribers)-1]
			pl.subscribers = pl.subscribers[0 : len(pl.subscribers)-1]
			break
		}
	}
	pl.mux.Unlock()
}

// Sends the playerlist to all subscribers. This function should be called any
// time the playerlist changes, while t.mux is still held.
func (pl *PlayerList) notifySubscribers() {
	playerlist := pl.playerlist()
	for _, channel := range pl.subscribers {
		// Use a non-blocking write in case the subscriber isn't draining
		// the channel.
		select {
		case channel <- playerlist:
		}
	}
}
