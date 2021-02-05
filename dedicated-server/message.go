package main

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

type messageType byte

const (
	// ConfigMessageType messages contain startup configuration information.
	// The payload of a config message from a client is the syncHash. The
	// payload of a config message from the server is
	// "<synchash>,<clientIu>,<ramconfig>".
	ConfigMessageType messageType = 'c'

	// KickPlayerMessageType messages indicate that the recipient should be
	// kicked. Only sent by the server. No payload.
	KickPlayerMessageType messageType = 'k'

	// PlayerListMessageType messages contain the names of all currently
	// connected players. The payload is a comma-separated list of
	// "l:<user>:num:<playerNum>,l:<user>:status:<status>" tuples.
	PlayerListMessageType messageType = 'l'

	// MemoryMessageType messages indicatesmodifications that should be made to
	// game memory. The payload is a comma-separated list of "<addr>:<value>"
	// pairs.
	MemoryMessageType messageType = 'm'

	// PlayerNumberMessageType messages contain the user's player number. The
	// payload is "<user>,<number>", where clients may omit "<number>" (and
	// possibly the leading comma) if they want the server to select their
	// player number.
	PlayerNumberMessageType messageType = 'n'

	// PingMessageType messages keep the connection alive. There is no payload.
	PingMessageType messageType = 'p'

	// QuitMessageType messages indicate the sender is disconnecting. The
	// payload is "q:" or "q:was_kicked" based on whether or not the player was
	// kicked.
	QuitMessageType messageType = 'q'

	// RAMEventMessageType messages contain ramcontroller-specific information.
	// The server sends the itemlist ("i:...") when new clients connect; after
	// that, all ram event messages are unintepreted by the server.
	RAMEventMessageType messageType = 'r'

	// PlayerStatusMessageType messages update the status of a player. The
	// payload is "<user>,<status>", where the status is a freeform string.
	PlayerStatusMessageType messageType = 's'
)

var (
	// ErrEmptyMessage is returned when attempting to decode an empty message.
	ErrEmptyMessage = errors.New("message is empty")
	// ErrUnknownMessageType is returned when attempting to decode a message
	// with an unknown type.
	ErrUnknownMessageType = errors.New("unsupported message type")
	// ErrMalformedMessage is returned whenever the message cannot be parsed.
	ErrMalformedMessage = errors.New("malformed message")
)

// Converts a messageType to a string.
func (t messageType) String() string {
	switch t {
	case ConfigMessageType:
		return "ConfigMessageType"
	case KickPlayerMessageType:
		return "KickPlayerMessageType"
	case PlayerListMessageType:
		return "PlayerListMessageType"
	case MemoryMessageType:
		return "MemoryMessageType"
	case PlayerNumberMessageType:
		return "PlayerNumberMessageType"
	case PingMessageType:
		return "PingMessageType"
	case QuitMessageType:
		return "QuitMessageType"
	case RAMEventMessageType:
		return "RAMEventMessageType"
	case PlayerStatusMessageType:
		return "PlayerStatusMessageType"
	default:
		return fmt.Sprintf("UNKNOWN(%c)", t)
	}
}

// Message contains all information sent in a client/server message.
type Message struct {
	// The type of the message.
	MessageType messageType
	// The name of the user from which the message was sent.
	FromUserName string
	// The payload of the message, which depends on the message type.
	Payload string
}

// DecodeMessage decodes a serialized message.
// The message format is "[type][fromUser],[payload]".
func DecodeMessage(serialized string) (*Message, error) {
	if len(serialized) == 0 {
		return nil, ErrEmptyMessage
	}

	// The first character of the message specifies its type.
	switch t := messageType(serialized[0]); t {
	case ConfigMessageType:
	case KickPlayerMessageType:
	case PlayerListMessageType:
	case MemoryMessageType:
	case PlayerNumberMessageType:
	case PingMessageType:
	case QuitMessageType:
	case RAMEventMessageType:
	case PlayerStatusMessageType:
	default:
		return nil, ErrUnknownMessageType
	}

	// The user name is before the first comma, and the payload follows it.
	parts := strings.SplitN(serialized[1:], ",", 2)
	if len(parts) != 2 {
		return nil, ErrMalformedMessage
	}
	return &Message{messageType(serialized[0]), parts[0], parts[1]}, nil
}

// Send writes the message to the Writer.
func (m Message) Send(w io.Writer) error {
	_, err := fmt.Fprintf(w, "%c%s,%s\n", m.MessageType, m.FromUserName, m.Payload)
	return err
}
