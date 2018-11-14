package util

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

// NewHTTPTunnel opens a connection to 127.0.0.1:8888 and initiates an HTTP
// tunnel by sending a CONNECT message.
func NewHTTPTunnel(serverName string) (net.Conn, error) {
	proxy, err := net.Dial("tcp", "127.0.0.1:8888")
	if err != nil {
		return nil, fmt.Errorf("error connecting to proxy: %v", err)
	}

	req := fmt.Sprintf("CONNECT %s:443 HTTP/1.1\r\n\r\n", serverName)
	_, err = proxy.Write([]byte(req))
	if err != nil {
		proxy.Close()
		return nil, fmt.Errorf("error sending CONNECT to proxy: %v", err)
	}

	var resp []byte
	reader := bufio.NewReader(proxy)
	for {
		line, isPrefix, err := reader.ReadLine()
		if err != nil {
			proxy.Close()
			return nil, fmt.Errorf("error reading proxy CONNECT response: %v", err)
		}

		resp = append(resp, line...)
		if !isPrefix {
			resp = append(resp, '\r', '\n')
		}

		if len(line) == 0 {
			break
		}
	}

	if !strings.HasPrefix(string(resp), "HTTP/1.1 200 ") {
		proxy.Close()
		return nil, fmt.Errorf("proxy CONNECT response not 200: %s", resp)
	}

	return proxy, nil
}
