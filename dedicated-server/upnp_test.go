package main

import (
	"errors"
	"fmt"
	"github.com/huin/goupnp"
	"net"
	"net/url"
	"reflect"
	"testing"
)

var ErrTest = errors.New("message")

type FakeWanConnection struct {
	addPortMapping       func(string, uint16, string, uint16, string, bool, string, uint32) error
	deletePortMapping    func(string, uint16, string) error
	getExternalIPAddress func() (string, error)
	getServiceClient     func() *goupnp.ServiceClient
}

func (wc *FakeWanConnection) AddPortMapping(
	remoteHost string, externalPort uint16, protocol string, internalPort uint16,
	internalClient string, enabled bool, portMappingDescription string, leaseDuration uint32) error {
	return wc.addPortMapping(
		remoteHost, externalPort, protocol, internalPort, internalClient, enabled,
		portMappingDescription, leaseDuration)
}

func (wc *FakeWanConnection) DeletePortMapping(
	remoteHost string, externalPort uint16, protocol string) error {
	return wc.deletePortMapping(remoteHost, externalPort, protocol)
}

func (wc *FakeWanConnection) GetExternalIPAddress() (string, error) {
	return wc.getExternalIPAddress()
}

func (wc *FakeWanConnection) GetServiceClient() *goupnp.ServiceClient {
	return wc.getServiceClient()
}

func TestFindRouterError(t *testing.T) {
	f := func() (wanConnection, error) { return nil, ErrTest }
	if _, err := newPortForwarder(f); !errors.Is(err, ErrTest) {
		t.Errorf("got %v, want %v", err, ErrTest)
	}
}

func TestGetExternalIPAddressFailure(t *testing.T) {
	wc := &FakeWanConnection{
		getExternalIPAddress: func() (string, error) { return "", ErrTest },
	}
	f := func() (wanConnection, error) { return wc, nil }
	if _, err := newPortForwarder(f); !errors.Is(err, ErrTest) {
		t.Errorf("got %v, want %v", err, ErrTest)
	}
}

func TestGetInternalIPFailure(t *testing.T) {
	sc := &goupnp.ServiceClient{
		Location: &url.URL{Host: "invalid"},
	}
	wc := &FakeWanConnection{
		getExternalIPAddress: func() (string, error) { return "", nil },
		getServiceClient:     func() *goupnp.ServiceClient { return sc },
	}
	f := func() (wanConnection, error) { return wc, nil }
	if _, err := newPortForwarder(f); err == nil {
		t.Errorf("getInternalIP with invalid host should fail")
	}
}

func TestExternalIP(t *testing.T) {
	conn, err := net.ListenPacket("udp", ":0")
	if err != nil {
		t.Errorf("listen failed: %v", err)
	}
	defer conn.Close()

	sc := &goupnp.ServiceClient{
		Location: &url.URL{Host: conn.LocalAddr().String()},
	}
	externalIP := "1.2.3.4"
	wc := &FakeWanConnection{
		getExternalIPAddress: func() (string, error) { return externalIP, nil },
		getServiceClient:     func() *goupnp.ServiceClient { return sc },
	}
	f := func() (wanConnection, error) { return wc, nil }
	pf, err := newPortForwarder(f)
	if err != nil {
		t.Errorf("newPortForwarder failed: %v", err)
	}
	if pf.ExternalIP != externalIP {
		t.Errorf("got %s, want %s", pf.ExternalIP, externalIP)
	}
}

func TestInternalIP(t *testing.T) {
	conn, err := net.ListenPacket("udp", ":0")
	if err != nil {
		t.Errorf("listen failed: %v", err)
	}
	defer conn.Close()

	sc := &goupnp.ServiceClient{
		Location: &url.URL{Host: conn.LocalAddr().String()},
	}
	wc := &FakeWanConnection{
		getExternalIPAddress: func() (string, error) { return "", nil },
		getServiceClient:     func() *goupnp.ServiceClient { return sc },
	}
	f := func() (wanConnection, error) { return wc, nil }
	pf, err := newPortForwarder(f)
	if err != nil {
		t.Errorf("newPortForwarder failed: %v", err)
	}
	internalIP := "127.0.0.1"
	if pf.InternalIP != internalIP {
		t.Errorf("got %s, want %s", pf.InternalIP, internalIP)
	}
}

func TestAddFailure(t *testing.T) {
	conn, err := net.ListenPacket("udp", ":0")
	if err != nil {
		t.Errorf("listen failed: %v", err)
	}
	defer conn.Close()

	sc := &goupnp.ServiceClient{
		Location: &url.URL{Host: conn.LocalAddr().String()},
	}
	wc := &FakeWanConnection{
		getExternalIPAddress: func() (string, error) { return "", nil },
		getServiceClient:     func() *goupnp.ServiceClient { return sc },
		addPortMapping: func(string, uint16, string, uint16, string, bool, string, uint32) error {
			return ErrTest
		},
	}
	f := func() (wanConnection, error) { return wc, nil }
	pf, err := newPortForwarder(f)
	if err != nil {
		t.Errorf("newPortForwarder failed: %v", err)
	}

	if err := pf.Add(0, 0, ""); !errors.Is(err, ErrTest) {
		t.Errorf("got %v, want %v", err, ErrTest)
	}
	// Close should be a no-op because Add failed.
	if err := pf.Close(); err != nil {
		t.Errorf("Close failed: %v", err)
	}
}

func TestCloseFailure(t *testing.T) {
	conn, err := net.ListenPacket("udp", ":0")
	if err != nil {
		t.Errorf("listen failed: %v", err)
	}
	defer conn.Close()

	sc := &goupnp.ServiceClient{
		Location: &url.URL{Host: conn.LocalAddr().String()},
	}
	deleteCalls := 0
	wc := &FakeWanConnection{
		getExternalIPAddress: func() (string, error) { return "", nil },
		getServiceClient:     func() *goupnp.ServiceClient { return sc },
		addPortMapping: func(string, uint16, string, uint16, string, bool, string, uint32) error {
			return nil
		},
		deletePortMapping: func(_ string, externalPort uint16, _ string) error {
			deleteCalls++
			return ErrTest
		},
	}
	f := func() (wanConnection, error) { return wc, nil }
	pf, err := newPortForwarder(f)
	if err != nil {
		t.Errorf("newPortForwarder failed: %v", err)
	}

	if err := pf.Add(100, 0, ""); err != nil {
		t.Errorf("Add failed: %v", err)
	}
	if err := pf.Add(200, 0, ""); err != nil {
		t.Errorf("Add failed: %v", err)
	}
	// Close should be a no-op because Add failed.
	if err := pf.Close(); !errors.Is(err, ErrTest) {
		t.Errorf("got %v, want %v", err, ErrTest)
	}
	if deleteCalls != 2 {
		t.Errorf("all ports not closed; got %v, want 2", deleteCalls)
	}
}

func TestPortMappings(t *testing.T) {
	conn, err := net.ListenPacket("udp4", ":0")
	if err != nil {
		t.Errorf("listen failed: %v", err)
	}
	defer conn.Close()

	sc := &goupnp.ServiceClient{
		Location: &url.URL{Host: conn.LocalAddr().String()},
	}
	var adds []string
	var deletes []string
	wc := &FakeWanConnection{
		getExternalIPAddress: func() (string, error) { return "", nil },
		getServiceClient:     func() *goupnp.ServiceClient { return sc },
		addPortMapping: func(
			remoteHost string, externalPort uint16, protocol string, internalPort uint16,
			internalClient string, enabled bool, portMappingDescription string, leaseDuration uint32) error {
			s := fmt.Sprintf(
				"%s,%d,%s,%d,%s,%t,%s,%d", remoteHost, externalPort, protocol, internalPort,
				internalClient, enabled, portMappingDescription, leaseDuration)
			adds = append(adds, s)
			return nil
		},
		deletePortMapping: func(remoteHost string, externalPort uint16, protocol string) error {
			deletes = append(deletes, fmt.Sprintf("%s,%d,%s", remoteHost, externalPort, protocol))
			return nil
		},
	}
	f := func() (wanConnection, error) { return wc, nil }
	pf, err := newPortForwarder(f)
	if err != nil {
		t.Errorf("newPortForwarder failed: %v", err)
	}

	if err := pf.Add(100, 1000, "100 -> 1000"); err != nil {
		t.Errorf("Add failed: %v", err)
	}
	if err := pf.Add(200, 2000, "200 -> 2000"); err != nil {
		t.Errorf("Add failed: %v", err)
	}
	wantAdds := []string{
		",100,TCP,1000,127.0.0.1,true,100 -> 1000,0",
		",200,TCP,2000,127.0.0.1,true,200 -> 2000,0",
	}
	if !reflect.DeepEqual(adds, wantAdds) {
		t.Errorf("got %v, want %v", adds, wantAdds)
	}

	if err := pf.Close(); err != nil {
		t.Errorf("Close failed: %v", err)
	}
	wantDeletes := []string{
		",100,TCP",
		",200,TCP",
	}
	if !reflect.DeepEqual(deletes, wantDeletes) {
		t.Errorf("got %v, want %v", deletes, wantDeletes)
	}
}
