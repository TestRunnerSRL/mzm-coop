package main

import (
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
)

type TestHandler struct {
	urls []string
}

func (h *TestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.urls = append(h.urls, r.URL.String())
}

func GetHost(ts *httptest.Server) string {
	u, err := url.Parse(ts.URL)
	if err != nil {
		log.Fatalf("failed to parse url: %v", err)
	}
	return u.Host
}

func TestRegisterRoom(t *testing.T) {
	th := &TestHandler{}
	ts := httptest.NewTLSServer(th)
	defer ts.Close()
	roomlistServerName = GetHost(ts)

	rr, err := RegisterRoom("Room", "password", ts.Client())
	if err != nil {
		t.Fatalf("registration failed: %v", err)
	}
	want := []string{"/create?pass=password&user=Room"}
	if !reflect.DeepEqual(th.urls, want) {
		t.Errorf("got %v, want %v", th.urls, want)
	}

	if err := rr.Close(); err != nil {
		t.Fatalf("unregistration failed: %v", err)
	}
	want = append(want, "/destroy?pass=password&user=Room")
	if !reflect.DeepEqual(th.urls, want) {
		t.Errorf("got %v, want %v", th.urls, want)
	}
}

func TestRegisterRoomFailure(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(http.NotFound))
	defer ts.Close()
	roomlistServerName = GetHost(ts)

	if _, err := RegisterRoom("Room", "password", ts.Client()); err == nil {
		t.Errorf("registration should have failed")
	}
}

func TestRegisterRoomConnectionFailure(t *testing.T) {
	ts := httptest.NewUnstartedServer(&TestHandler{})
	defer ts.Close()
	roomlistServerName = GetHost(ts)

	if _, err := RegisterRoom("Room", "password", http.DefaultClient); err == nil {
		t.Errorf("registration should have failed")
	}
}

func TestCloseRegisteredRoomFailure(t *testing.T) {
	status := http.StatusOK
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
	}))
	defer ts.Close()
	roomlistServerName = GetHost(ts)

	rr, err := RegisterRoom("Room", "password", ts.Client())
	if err != nil {
		t.Fatalf("registration failed: %v", err)
	}

	status = http.StatusNotFound
	if err := rr.Close(); err == nil {
		t.Errorf("unregistration should have failed")
	}
}

func TestCloseRegisteredRoomConnectionFailure(t *testing.T) {
	ts := httptest.NewTLSServer(&TestHandler{})
	roomlistServerName = GetHost(ts)

	rr, err := RegisterRoom("Room", "password", ts.Client())
	if err != nil {
		ts.Close()
		t.Fatalf("registration failed: %v", err)
	}

	ts.Close()
	if err := rr.Close(); err == nil {
		t.Errorf("unregistration should have failed")
	}
}
