package main

import (
	"net"
	"time"
)

// tcpKeepAliveListener imposta i timeout keepalive TCP sulle connessioni accettate.
// Viene utilizzato da ListenAndServe e ListenAndServeTLS quindi connessioni TCP inattive
// (es. chiusura del laptop durante il download) alla fine scompare. Questo è il codice di
// net / http / server.go.
type tcpKeepAliveListener struct {
	*net.TCPListener
}

// Accept accetta una connessione TCP durante l'impostazione dei timeout keep-alive.
func (ln tcpKeepAliveListener) Accept() (c net.Conn, err error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(3 * time.Minute)
	return tc, nil
}