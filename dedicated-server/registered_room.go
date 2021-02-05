// Allows creation and deletion of registered BizHawk co-op rooms.

package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

// Name of the registration server. This is not const to allow injection
// in unit tests.
var roomlistServerName = "us-central1-mzm-coop.cloudfunctions.net"

// RegisteredRoom enables registering and unregistering rooms on the BizHawk
// co-op server.
type RegisteredRoom struct {
	name, pass string
	client     *http.Client
}

// RegisterRoom registers a new room with the provided name and password. The
// room should be closed when no longer needed.
func RegisterRoom(name string, pass string, client *http.Client) (*RegisteredRoom, error) {
	r := &RegisteredRoom{name, pass, client}
	if err := r.sendCommand("/create"); err != nil {
		return nil, fmt.Errorf("failed to register room: %w", err)
	}
	return r, nil
}

// Close unregisters the room.
func (r *RegisteredRoom) Close() error {
	if err := r.sendCommand("/destroy"); err != nil {
		return fmt.Errorf("failed to unregister room: %w", err)
	}
	return nil
}

// sendCommand sends the given command to the roomlist server.
func (r *RegisteredRoom) sendCommand(command string) error {
	v := url.Values{}
	v.Set("user", r.name)
	v.Set("pass", r.pass)
	u := url.URL{
		Scheme:   "https",
		Host:     roomlistServerName,
		Path:     command,
		RawQuery: v.Encode(),
	}

	resp, err := r.client.Get(u.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return errors.New(resp.Status)
	}
	return nil
}
