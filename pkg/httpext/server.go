package httpext

import (
	"net"
	"net/http"
)

func NewServer(handler http.Handler) (url string, server *http.Server, l net.Listener, retErr error) {
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return "", nil, listener, err
	}

	addr := listener.Addr().String()
	server = &http.Server{
		Addr:    addr,
		Handler: handler,
	}
	url = "http://" + addr
	return url, server, listener, nil
}
