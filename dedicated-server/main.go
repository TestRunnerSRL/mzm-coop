package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const usage = `USAGE: dedicated_server [FLAGS]

Runs a BizHawk co-op server compatible with the in-emulator
Lua scrips. Running a dedicated server has better performance
with a large number of clients, and ensures that the co-op
game is not dependent on the host's BizHawk process. If
necessary, the dedicated server can be run on a completely
different server (e.g., AWS EC2 or GCP).

`

func addPortForwarding(externalPort int, internalPort int) (string, string, io.Closer, error) {
	pf, err := NewPortForwarder()
	if err != nil {
		return "", "", nil, err
	}
	if err := pf.Add(uint16(externalPort), uint16(internalPort), "BizHawk co-op"); err != nil {
		return "", "", nil, err
	}
	return pf.InternalIP, pf.ExternalIP, pf, nil
}

func main() {
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, usage)
		flag.PrintDefaults()
	}
	var license = flag.Bool("license", false, "Prints license information for this binary and its dependencies.")
	var port = flag.Int("port", 50000, "TCP/IP port on which the server runs.")
	var upnpPort = flag.Int("upnpport", 0, "If non-zero, enables port forwarding from this external port using UPnP.")
	var adminPort = flag.Int("adminport", 8080, "If non-zero, a web-based admin interface will run on this port.")
	var roomName = flag.String("room", "", "If non-empty, registers a room with this name.")
	var pass = flag.String("password", "", "Password for the registered room.")
	var syncHash = flag.String("synchash", "", "Configuration hash to ensure consistent versions. If empty, the config will be inferred from the first client to connect.")
	var ramConfig = flag.String("ramconfig", "", "Game-specific configuration string.")
	var itemCount = flag.Int("itemcount", 1, "Number of items supported by the game.")
	flag.Parse()
	if *license {
		fmt.Fprint(os.Stderr, Licenses)
		return
	}

	rand.Seed(time.Now().UnixNano())

	// Create the server, which should be closed on shutdown.
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	// If *port is 0, the system will pick an available port.
	actualPort := listener.Addr().(*net.TCPAddr).Port

	// Set up port forwarding, and tear it down on exit.
	hostPort := fmt.Sprintf("localhost:%d", actualPort)
	if *upnpPort > 0 {
		log.Print("Setting up port forwarding...")
		internalIP, externalIP, closer, err := addPortForwarding(*upnpPort, actualPort)
		if err != nil {
			log.Printf("Port forwarding failed: %v", err)
		} else {
			hostPort = fmt.Sprintf("%s:%d (internal) and %s:%d (external)", internalIP, actualPort, externalIP, *upnpPort)
			defer func() {
				if err := closer.Close(); err != nil {
					log.Printf("Failed to remove port forwarding: %v", err)
				}
			}()
		}
	}

	// If --room is set, register the room so that other users can find it.
	if *roomName != "" {
		rr, err := RegisterRoom(*roomName, *pass, http.DefaultClient)
		if err != nil {
			log.Print(err)
		}
		log.Printf("Server registered as %s", *roomName)
		defer rr.Close()
	}

	// Create the room, which should be closed on shutdown.
	pl := NewPlayerList()
	room := NewRoom(NewSyncConfig(*syncHash, *ramConfig, *itemCount), pl)
	if *adminPort > 0 {
		log.Printf("Admin UI: http://localhost:%d", *adminPort)
		admin := NewAdminServer(*adminPort, room, pl)
		defer admin.Close()
	}

	// Run until the listener is closed due to the listener being closed.
	log.Printf("Running on %s", hostPort)
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				// An error usually indicates the listener was closed.
				break
			}
			go room.HandleConnection(conn)
		}
	}()

	// Wait until SIGINT and SIGTERM is received.
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	log.Printf("Shutting down...")
	// First close the listener so that new connections are not accepted.
	if err := listener.Close(); err != nil {
		log.Printf("Server shutdown failed: %v", err)
	}
	if err := room.Close(); err != nil {
		log.Printf("Failed to disconnect users: %v", err)
	}
}
