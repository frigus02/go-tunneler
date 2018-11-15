package main

import (
	"flag"
	"fmt"
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

func handleConnection(c net.Conn, proxyAddress string) {
	tag := c.RemoteAddr().String()
	log.Printf("[%s] New connection\n", tag)

	serverName, readBytes, err := util.GetServerName(c)
	if err != nil {
		log.Printf("[%s] Error getting server name: %v\n", tag, err)
		c.Close()
		return
	}

	log.Printf("[%s] Found server name %s\n", tag, serverName)

	proxy, err := util.NewHTTPTunnel(proxyAddress, serverName)
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
	var proxyAddress string
	var port int
	var help bool
	flag.StringVar(&proxyAddress, "proxy", "127.0.0.1:3128", "Proxy address")
	flag.IntVar(&port, "port", 443, "Port the Go Tunneler should listen on")
	flag.BoolVar(&help, "help", false, "Show help")
	flag.Parse()

	if help {
		flag.Usage()
		return
	}

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Printf("Error listening on 8443: %v\n", err)
		return
	}

	defer l.Close()

	log.Printf("Listening on %d, waiting for connections...\n", port)
	for {
		c, err := l.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v\n", err)
			return
		}

		go handleConnection(c, proxyAddress)
	}
}
