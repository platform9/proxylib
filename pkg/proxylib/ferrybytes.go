package proxylib

import (
	"io"
	"log"
	"net"
	"strings"
	"time"
)

const (
	ConnectionClosedErr = "use of closed network connection"
	ConnectionResetErr  = "connection reset by peer"
	DefaultMaxTeardownTimeInSeconds = 10
)

// Ferry bytes between two sockets.
// Tries to handle disconnections safely.
// When one socket disconnects, a CloseWrite() is done to the other socket, and
// a teardown timer started with the specified timeout value (if value is zero,
// a default value of 10 seconds is used). If a graceful disconnect from the
// other socket is not detected within that period, the other end is
// forcefully closed.
func FerryBytes(
	client *net.TCPConn,
	server *net.TCPConn,
	cnxId string,
	maxTeardownTimeInSecs int,
) {
	log.Printf("[%s] Initiating copy between %s and %s", cnxId,
		client.RemoteAddr().String(), server.RemoteAddr().String())

	doCopy := func(s, c *net.TCPConn, cancel chan<- string) {
		numWritten, err := io.Copy(s, c)
		reason := "EOF"
		if err != nil {
			reason = err.Error()
		}
		log.Printf("[%s] Copied %d bytes from %s to %s, finished because: %s",
			cnxId, numWritten, c.RemoteAddr().String(),
			s.RemoteAddr().String(),
			reason)
		if err != nil && !strings.Contains(err.Error(),
			ConnectionClosedErr) && !strings.Contains(err.Error(),
				ConnectionResetErr) {
			log.Printf("[%s] Failed copying connection data: %v",
				cnxId, err)
		}
		log.Printf("[%s] Copy finished for %s -> %s", cnxId,
			c.RemoteAddr().String(), s.RemoteAddr().String())
		err = s.CloseWrite() // propagate EOF signal to destination
		if err != nil {
			log.Printf("[%s] warning: failed to CloseWrite() %s -> %s : %s --ok",
				cnxId, c.RemoteAddr().String(), s.RemoteAddr().String(), err)
		}
		cancel <- c.RemoteAddr().String()
	}

	cancel := make(chan string, 2)
	go doCopy(server, client, cancel)
	go doCopy(client, server, cancel)

	closedSrc := <- cancel
	log.Printf("[%s] 1st source to close: %s", cnxId, closedSrc)
	if maxTeardownTimeInSecs == 0 {
		maxTeardownTimeInSecs = DefaultMaxTeardownTimeInSeconds
	}
	timer := time.NewTimer(time.Duration(maxTeardownTimeInSecs) * time.Second)
	select {
	case closedSrc = <-cancel:
		log.Printf("[%s] 2nd source to close: %s (all done)",
			cnxId, closedSrc)
		timer.Stop()
	case <- timer.C:
		log.Printf("[%s] timed out waiting for 2nd source to close",
			cnxId)
	}
}
