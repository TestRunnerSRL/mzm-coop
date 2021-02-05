// A loadtest for BizHawk Co-Op servers that simulates a potentially very large
// number of clients connecting, sending messages, and disconnecting. The
// loadtest collects and displays both latency and error rate metrics.
//
// Example:
//
//   go run loadtest.go --synchash=<HASH>
//
// +build ignore

package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"sort"
	"strings"
	"sync/atomic"
	"syscall"
	"time"
)

type FakeClient struct {
	// The "host:port" to which the client connects.
	host string
	// The name of the client (e.g., "Player1").
	name string
	// The syncHash used when connecting to the server.
	syncHash string
	// The average time before the client disconnects.
	dcTime time.Duration
	// The average time between sending memory messages.
	msgPeriod time.Duration
	// The number of memory messages the client has sent.
	msgCount int
	// The StatsCollector to which metrics are reported.
	stats *StatsCollector
}

func NewFakeClient(port int, name string, syncHash string, dcTime time.Duration, msgPeriod time.Duration, stats *StatsCollector) *FakeClient {
	return &FakeClient{
		host:      fmt.Sprintf("localhost:%d", port),
		name:      name,
		syncHash:  syncHash,
		dcTime:    dcTime,
		msgPeriod: msgPeriod,
		stats:     stats,
	}
}

func (c *FakeClient) Run() error {
	conn, err := net.Dial("tcp", c.host)
	if err != nil {
		return err
	}
	defer conn.Close()
	go c.ping(conn)
	go c.handleInput(conn)

	// Clients send their config and player number as soon as they connect.
	if _, err := fmt.Fprintf(conn, "c%s,%s\n", c.name, c.syncHash); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(conn, "n%s,%s,\n", c.name, c.name); err != nil {
		return err
	}

	// Periodically send memory messages until it's time to disconnect.
	memory := time.NewTimer(c.randomDuration(c.msgPeriod))
	disconnect := time.NewTimer(c.randomDuration(c.dcTime))
	for {
		select {
		case <-memory.C:
			c.msgCount++
			msg := fmt.Sprintf("r%s,n:%d\n", c.name, c.msgCount)
			c.stats.RecordMessageSent(msg[0:len(msg)-1], time.Now())
			if _, err := conn.Write([]byte(msg)); err != nil {
				return err
			}
			memory.Reset(c.randomDuration(c.msgPeriod))
		case <-disconnect.C:
			if !memory.Stop() {
				<-memory.C
			}
			return nil
		}
	}
}

// Returns a random duration value from an exponential distribution with the
// given average.
func (c *FakeClient) randomDuration(avg time.Duration) time.Duration {
	return time.Duration(float64(avg) * rand.ExpFloat64())
}

// Reads input from the connection and records metrics.
func (c *FakeClient) handleInput(conn net.Conn) {
	start := time.Now()
	lastPing := time.Time{}
	serverName := ""
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		switch scanner.Text()[0] {
		case 'c':
			if !start.IsZero() {
				c.stats.RecordDuration("connect", time.Now().Sub(start))
				start = time.Time{}
				serverName = strings.SplitN(scanner.Text()[1:], ",", 2)[0]
			}
		case 'r':
			c.stats.RecordMessageReceipt(scanner.Text(), time.Now())
		case 'p':
			// Ignore pings from anyone other than the server.
			if strings.HasPrefix(scanner.Text()[1:], serverName) {
				now := time.Now()
				if !lastPing.IsZero() {
					c.stats.RecordDuration("ping", now.Sub(lastPing))
				}
				lastPing = now
			}
		}
	}
}

// Keeps the connection alive by sending pings every 10 seconds.
func (c *FakeClient) ping(conn net.Conn) {
	ticker := time.NewTicker(10 * time.Second)
	msg := []byte(fmt.Sprintf("p%s,\n", c.name))
	for {
		<-ticker.C
		if _, err := conn.Write(msg); err != nil {
			// The connection was probably closed; stop pinging.
			break
		}
	}
	ticker.Stop()
}

// Collects and periodically prints statistics.
type StatsCollector struct {
	messages map[string]time.Time
	records  map[string][]float64
	alpha    float64
	ema      map[string]float64
	emv      map[string]float64
	channel  chan interface{}
	done     chan struct{}
}

type messageSentObservation struct {
	msg string
	t   time.Time
}

type messageReceivedObservation struct {
	msg string
	t   time.Time
}

type durationObservation struct {
	name string
	d    time.Duration
}

// Creates a new StatsCollector that reports exponentially windowed statistics
// over (roughly) the last N observations.
func NewStatsCollector(n int) *StatsCollector {
	sc := &StatsCollector{
		messages: make(map[string]time.Time),
		records:  make(map[string][]float64),
		alpha:    2 / (float64(n) + 1),
		ema:      make(map[string]float64),
		emv:      make(map[string]float64),
		channel:  make(chan interface{}, 100),
		done:     make(chan struct{}, 1),
	}
	go sc.run()
	return sc
}

// Records that a message was sent. Subsequent calls to RecordMessageReceipt()
// will cause a duration to be recorded.
func (sc *StatsCollector) RecordMessageSent(msg string, t time.Time) {
	sc.channel <- messageSentObservation{msg, t}
}

// Records that a message was received. This is equivalent to calling
// RecordDuration() with the amount of time between the message being
// sent and received.
func (sc *StatsCollector) RecordMessageReceipt(msg string, t time.Time) {
	sc.channel <- messageReceivedObservation{msg, t}
}

// Records an observation of a duration to a metric.
func (sc *StatsCollector) RecordDuration(name string, d time.Duration) {
	sc.channel <- durationObservation{name, d}
}

// Stops collecting new metrics and reports statistics on all metrics
// collected thus far.
func (sc *StatsCollector) Close() error {
	sc.done <- struct{}{}
	// Wait for the metrics to be printed.
	<-sc.done
	return nil
}

// The StatsCollector's main gofunction, which periodically prints
// exponentially weighted statistics until the StatsCollector is closed.
func (sc *StatsCollector) run() {
	printStats := time.Tick(5 * time.Second)
	for keepRunning := true; keepRunning; {
		select {
		case x := <-sc.channel:
			if mso, ok := x.(messageSentObservation); ok {
				sc.messages[mso.msg] = mso.t
			} else if mro, ok := x.(messageReceivedObservation); ok {
				if start, ok := sc.messages[mro.msg]; ok {
					sc.recordObservation("message", float64(mro.t.Sub(start)))
				}
			} else if do, ok := x.(durationObservation); ok {
				sc.recordObservation(do.name, float64(do.d))
			}
		case <-printStats:
			descriptions := make([]string, 0, len(sc.ema))
			for name, ema := range sc.ema {
				desc := fmt.Sprintf("%s: %v ± %v", name, time.Duration(ema), time.Duration(math.Sqrt(sc.emv[name])))
				descriptions = append(descriptions, desc)
			}
			if len(descriptions) > 0 {
				sort.Strings(descriptions)
				log.Print(strings.Join(descriptions, "; "))
			}

			// Also clean up old messages.
			threshold := time.Now().Add(-1 * time.Minute)
			for msg, t := range sc.messages {
				if t.Before(threshold) {
					delete(sc.messages, msg)
				}
			}
		case <-sc.done:
			keepRunning = false
		}
	}

	// When exiting, record percentile information.
	for name, records := range sc.records {
		sort.Float64s(records)
		count := len(records)
		p := func(x float64) time.Duration {
			return time.Duration(records[int(x*float64(len(records)))])
		}
		log.Printf("%s (%d samples): 50p %v, 95p %v, 99p %v", name, count, p(0.5), p(0.95), p(0.99))
	}
	close(sc.done)
}

// Records an observation and updates the exponentially weighted stats.
func (sc *StatsCollector) recordObservation(name string, x float64) {
	// Record that the value for percentile computation later.
	sc.records[name] = append(sc.records[name], x)

	// Compute the exponentially weighted moving average and variance as per
	// http://en.wikipedia.org/wiki/Moving_average#Exponentially_weighted_moving_variance_and_standard_deviation.
	if ema, ok := sc.ema[name]; ok {
		delta := x - ema
		sc.ema[name] = ema + sc.alpha*delta
		sc.emv[name] = (1 - sc.alpha) * (sc.emv[name] + sc.alpha*delta*delta)
	} else {
		sc.ema[name] = x
		sc.emv[name] = 0
	}
}

func main() {
	var clients = flag.Int("clients", 10, "Number of clients to simulate.")
	var port = flag.Int("port", 50000, "TCP/IP port on which the server runs.")
	var syncHash = flag.String("synchash", "", "Configuration hash to ensure consistent versions.")
	var dcPeriod = flag.Int("dcperiod", 300, "The average time in seconds between disconnects (per client).")
	var msgPeriod = flag.Int("msgperiod", 20, "The average time in seconds between sending messages (per client).")
	flag.Parse()

	rand.Seed(time.Now().UnixNano())
	stats := NewStatsCollector(20)
	defer stats.Close()
	dcPeriodDuration := time.Duration(*dcPeriod) * time.Second
	msgPeriodDuration := time.Duration(*msgPeriod) * time.Second

	log.Printf("Starting loadtest with %d clients", *clients)
	startTime := time.Now()
	var errors int64
	for i := 0; i < *clients; i++ {
		name := fmt.Sprintf("P%d", i)
		go func() {
			client := NewFakeClient(*port, name, *syncHash, dcPeriodDuration, msgPeriodDuration, stats)
			for {
				if err := client.Run(); err != nil {
					log.Print(err)
					atomic.AddInt64(&errors, 1)
				}
			}
		}()
	}

	// Wait until SIGINT or SIGTERM is received.
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	log.Printf("Shutting down...")
	errorRate := float64(atomic.LoadInt64(&errors)) * float64(time.Second) / float64(time.Now().Sub(startTime))
	log.Printf("errors: %f/sec", errorRate)
}
