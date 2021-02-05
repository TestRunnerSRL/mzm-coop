package main

import (
	"bytes"
	"errors"
	"testing"
)

func TestMessageTypeString(t *testing.T) {
	var tests = []struct {
		mt   messageType
		want string
	}{
		{ConfigMessageType, "ConfigMessageType"},
		{KickPlayerMessageType, "KickPlayerMessageType"},
		{PlayerListMessageType, "PlayerListMessageType"},
		{MemoryMessageType, "MemoryMessageType"},
		{PlayerNumberMessageType, "PlayerNumberMessageType"},
		{PingMessageType, "PingMessageType"},
		{QuitMessageType, "QuitMessageType"},
		{RAMEventMessageType, "RAMEventMessageType"},
		{PlayerStatusMessageType, "PlayerStatusMessageType"},
		{'x', "UNKNOWN(x)"},
	}

	for _, tt := range tests {
		if ans := tt.mt.String(); ans != tt.want {
			t.Errorf("got %s, want %s", ans, tt.want)
		}
	}
}

func TestDecodeMessage(t *testing.T) {
	var tests = []struct {
		encoded string
		want    Message
	}{
		{"cUser,Payload", Message{ConfigMessageType, "User", "Payload"}},
		{"kUser,Payload", Message{KickPlayerMessageType, "User", "Payload"}},
		{"lUser,Payload", Message{PlayerListMessageType, "User", "Payload"}},
		{"mUser,Payload", Message{MemoryMessageType, "User", "Payload"}},
		{"nUser,Payload", Message{PlayerNumberMessageType, "User", "Payload"}},
		{"pUser,", Message{PingMessageType, "User", ""}},
		{"qUser,", Message{QuitMessageType, "User", ""}},
		{"rUser,Payload", Message{RAMEventMessageType, "User", "Payload"}},
		{"sUser,Payload", Message{PlayerStatusMessageType, "User", "Payload"}},
	}

	for _, tt := range tests {
		ans, err := DecodeMessage(tt.encoded)
		if err != nil {
			t.Errorf("decoding %s failed: %v", tt.encoded, err)
		}
		if *ans != tt.want {
			t.Errorf("got %s, want %s", ans, tt.want)
		}
	}
}

func TestDecodeMessageFailure(t *testing.T) {
	var tests = []struct {
		encoded string
		want    error
	}{
		{"", ErrEmptyMessage},
		{"xUser,Payload", ErrUnknownMessageType},
		{"cNoPayload", ErrMalformedMessage},
	}

	for _, tt := range tests {
		if _, err := DecodeMessage(tt.encoded); !errors.Is(err, tt.want) {
			t.Errorf("got %v, want %v", err, tt.want)
		}
	}
}

func TestSend(t *testing.T) {
	buf := new(bytes.Buffer)
	msg := Message{MemoryMessageType, "User", "Payload"}
	expected := "mUser,Payload\n"
	if err := msg.Send(buf); err != nil {
		t.Fatalf("unexpected Send failure: %s", err)
	} else if buf.String() != expected {
		t.Fatalf("got %s, expected %s", buf.String(), expected)
	}
}
