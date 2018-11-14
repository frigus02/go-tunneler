package main

import (
	"io"
	"log"
	"net"

	"github.com/frigus02/go-tunneler/util"
)

func copyAndClose(dst io.WriteCloser, src io.Reader, logTag string) {
	_, err := io.Copy(dst, src)
	if err != nil {
		log.Printf("[%s] Error copying from src to dst: %v\n", logTag, err)
	}

	dst.Close()
}

func handleConnection(c net.Conn) {
	tag := c.RemoteAddr().String()
	log.Printf("[%s] New connection\n", tag)

	serverName, readBytes, err := util.GetServerName(c)
	if err != nil {
		log.Printf("[%s] Error getting server name: %v\n", tag, err)
		c.Close()
		return
	}

	log.Printf("[%s] Found server name %s\n", tag, serverName)

	proxy, err := util.NewHTTPTunnel(serverName)
	if err != nil {
		log.Printf("[%s] Error connecting to proxy: %v\n", tag, err)
		c.Close()
		return
	}

	log.Printf("[%s] Connected to proxy\n", tag)

	_, err = proxy.Write(readBytes)
	if err != nil {
		log.Printf("[%s] Error sending initial TLS handshake bytes through proxy: %v\n", tag, err)
		proxy.Close()
		c.Close()
		return
	}

	go copyAndClose(proxy, c, tag)
	go copyAndClose(c, proxy, tag)
}

func main() {
	l, err := net.Listen("tcp", ":8443")
	if err != nil {
		log.Printf("Error listening on 8443: %v\n", err)
		return
	}

	defer l.Close()

	log.Printf("Listening on 8443, waiting for connections...\n")
	for {
		c, err := l.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v\n", err)
			return
		}

		go handleConnection(c)
	}
}
