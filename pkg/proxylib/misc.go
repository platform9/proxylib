package proxylib

import (
	"io"
	"log"
)

func CloseConnection(cnx io.Closer, id string, name string) {
	if err := cnx.Close(); err != nil {
		log.Printf("[%s] warning: failed to close %s connection: %s",
			id, name, err)
	}
}
