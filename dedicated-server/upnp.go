package main

import (
	"errors"
	"github.com/huin/goupnp"
	"github.com/huin/goupnp/dcps/internetgateway2"
	"net"
)

// Satisfied by both internetgateway1.WANIPConnection1 and
// internetgateway1.WANPPPConnection1.
type wanConnection interface {
	AddPortMapping(
		remoteHost string, externalPort uint16, protocol string, internalPort uint16,
		internalClient string, enabled bool, portMappingDescription string,
		leaseDuration uint32) error
	DeletePortMapping(remoteHost string, externalPort uint16, protocol string) error
	GetExternalIPAddress() (string, error)
	GetServiceClient() *goupnp.ServiceClient
}

// TCPPortForwarder uses UPnP to enable port forwarding until closed.
type TCPPortForwarder struct {
	// The router's external IP address.
	ExternalIP string
	InternalIP string
	client     wanConnection
	ports      []uint16
}

// Finds a router as a WANIPConnection1, WANIPConnection2, or WANPPPConnection1
// device. In case there are multiple, the first is returned arbitrarily.
func findRouter() (wanConnection, error) {
	if clients, _, _ := internetgateway2.NewWANIPConnection1Clients(); len(clients) > 0 {
		return clients[0], nil
	}
	if clients, _, _ := internetgateway2.NewWANIPConnection2Clients(); len(clients) > 0 {
		return clients[0], nil
	}
	if clients, _, _ := internetgateway2.NewWANPPPConnection1Clients(); len(clients) > 0 {
		return clients[0], nil
	}
	return nil, errors.New("no UPnP routers found")
}

// Gets the internal IP address by connecting to the router and checking the
// address of the connection used.
func getInternalIPAddress(wc wanConnection) (string, error) {
	conn, err := net.Dial("udp", wc.GetServiceClient().Location.Host)
	if err != nil {
		return "", err
	}
	defer conn.Close()
	return conn.LocalAddr().(*net.UDPAddr).IP.String(), nil
}

// NewPortForwarder constructs a new PortForwarder and enables port forwarding.
func NewPortForwarder() (*TCPPortForwarder, error) {
	return newPortForwarder(findRouter)
}

// Internal implementation of NewPortForwarded that `findRouter` to be faked.
func newPortForwarder(f func() (wanConnection, error)) (*TCPPortForwarder, error) {
	client, err := f()
	if err != nil {
		return nil, err
	}

	externalIP, err := client.GetExternalIPAddress()
	if err != nil {
		return nil, err
	}

	internalIP, err := getInternalIPAddress(client)
	if err != nil {
		return nil, err
	}

	return &TCPPortForwarder{externalIP, internalIP, client, nil}, nil
}

// Add enables port forwarding from the external port to the internal port.
func (pf *TCPPortForwarder) Add(externalPort uint16, internalPort uint16, desc string) error {
	if err := pf.client.AddPortMapping("", externalPort, "TCP", internalPort, pf.InternalIP, true, desc, 0); err != nil {
		return err
	}
	pf.ports = append(pf.ports, externalPort)
	return nil
}

// Close tears down the port forwarding.
func (pf *TCPPortForwarder) Close() error {
	// Return the first error encountered, but still try to delete all port
	// mappings.
	var err error
	for _, port := range pf.ports {
		e := pf.client.DeletePortMapping("", port, "TCP")
		if err == nil && e != nil {
			err = e
		}
	}
	return err
}
