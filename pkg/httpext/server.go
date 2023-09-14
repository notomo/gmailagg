package httpext

import (
	"fmt"
	"net"
	"net/http"
)

func NewServer(
	handler http.Handler,
	port uint,
) (url string, server *http.Server, l net.Listener, retErr error) {
	host := fmt.Sprintf("localhost:%d", port)
	listener, err := net.Listen("tcp", host)
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
