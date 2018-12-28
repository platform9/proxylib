package proxylib

import (
	"io"
)

type Logger interface {
	Printf(format string, v ...interface{})
}

func CloseConnection(cnx io.Closer, log Logger, id string, name string) {
	log.Printf("[%s] closing %s connection", id, name)
	if err := cnx.Close(); err != nil {
		log.Printf("[%s] warning: failed to close %s connection: %s",
			id, name, err)
	}
}
