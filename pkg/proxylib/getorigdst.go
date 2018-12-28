package proxylib

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"syscall"
	"unsafe"
)

type sockaddr struct {
	family uint16
	data   [14]byte
}

const SO_ORIGINAL_DST = 80

var logger = log.New(os.Stderr, "", log.LstdFlags)

// OriginalDestination returns an intercepted connection's original destination.
// Adapted from https://play.golang.org/p/GMAaKucHOr
func OriginalDestination(cnxId string, conn *net.Conn) (string, error) {
	tcpConn, ok := (*conn).(*net.TCPConn)
	if !ok {
		return "", errors.New("not a TCPConn")
	}

	/*
	file, err := tcpConn.File()
	if err != nil {
		return "", fmt.Errorf("failed to get file from TCP connection: %s", err)
	}

	// To avoid potential problems from making the socket non-blocking.
	err = tcpConn.Close()
	if err != nil {
		return "", fmt.Errorf("failed to close TCP connection: %s", err)
	}
	*conn, err = net.FileConn(file)
	if err != nil {
		return "", fmt.Errorf("failed to get file connection: %s", err)
	}

	defer file.Close()
	fd := file.Fd()
	*/

	file, err := tcpConn.File()
	if err != nil {
		return "", fmt.Errorf("failed to get file from TCP connection: %s", err)
	}
	defer CloseConnection(file, logger, cnxId, "file")
	fd := file.Fd()

	var addr sockaddr
	size := uint32(unsafe.Sizeof(addr))
	err = getsockopt(
		int(fd), syscall.SOL_IP, SO_ORIGINAL_DST,
		uintptr(unsafe.Pointer(&addr)), &size,
	)
	if err != nil {
		return "", fmt.Errorf("getsockopt() failed: %s", err)
	}

	var ip net.IP
	switch addr.family {
	case syscall.AF_INET:
		ip = addr.data[2:6]
	default:
		return "", errors.New("unrecognized address family")
	}

	port := int(addr.data[0])<<8 + int(addr.data[1])
	return net.JoinHostPort(ip.String(), strconv.Itoa(port)), nil
}

func getsockopt(
	s int,
	level int,
	name int,
	val uintptr,
	vallen *uint32,
) (err error) {
	_, _, e1 := syscall.Syscall6(
		syscall.SYS_GETSOCKOPT, uintptr(s),
		uintptr(level), uintptr(name), uintptr(val),
		uintptr(unsafe.Pointer(vallen)), 0,
	)
	if e1 != 0 {
		err = e1
	}
	return
}
