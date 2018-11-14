package util

import (
	"fmt"
	"net"
)

// GetServerName extracts the server name from the SNI extension of a
// ClientHello message of a TLS connection.
// Credits: https://github.com/gpjt/stupid-proxy/blob/4f0e697f574d043734814c67a021a39b878aea93/proxy.go
func GetServerName(c net.Conn) (serverName string, readBytes []byte, err error) {
	firstByte := make([]byte, 1)
	_, err = c.Read(firstByte)
	if err != nil {
		return "", nil, fmt.Errorf("error reading first byte: %v", err)
	}
	if firstByte[0] != 0x16 {
		return "", nil, fmt.Errorf("not tls")
	}

	versionBytes := make([]byte, 2)
	_, err = c.Read(versionBytes)
	if err != nil {
		return "", nil, fmt.Errorf("error reading version bytes: %v", err)
	}
	if versionBytes[0] < 3 || (versionBytes[0] == 3 && versionBytes[1] < 1) {
		return "", nil, fmt.Errorf("SSL < 3.1 so still not TLS")
	}

	restLengthBytes := make([]byte, 2)
	_, err = c.Read(restLengthBytes)
	if err != nil {
		return "", nil, fmt.Errorf("error reading rest length bytes: %v", err)
	}
	restLength := (int(restLengthBytes[0]) << 8) + int(restLengthBytes[1])

	rest := make([]byte, restLength)
	_, err = c.Read(rest)
	if err != nil {
		return "", nil, fmt.Errorf("error reading rest bytes: %v", err)
	}

	current := 0

	handshakeType := rest[current]
	current++
	if handshakeType != 0x1 {
		return "", nil, fmt.Errorf("not a ClientHello")
	}

	// Skip over another length
	current += 3

	// Skip over protocolversion
	current += 2

	// Skip over random number
	current += 4 + 28

	// Skip over session ID
	sessionIDLength := int(rest[current])
	current++
	current += sessionIDLength

	cipherSuiteLength := (int(rest[current]) << 8) + int(rest[current+1])
	current += 2
	current += cipherSuiteLength

	compressionMethodLength := int(rest[current])
	current++
	current += compressionMethodLength
	if current > restLength {
		return "", nil, fmt.Errorf("no extensions")
	}

	// Skip over extensionsLength
	// extensionsLength := (int(rest[current]) << 8) + int(rest[current + 1])
	current += 2

	for current < restLength && serverName == "" {
		extensionType := (int(rest[current]) << 8) + int(rest[current+1])
		current += 2

		extensionDataLength := (int(rest[current]) << 8) + int(rest[current+1])
		current += 2

		if extensionType == 0 {
			// Skip over number of names as we're assuming there's just one
			current += 2

			nameType := rest[current]
			current++
			if nameType != 0 {
				return "", nil, fmt.Errorf("not a hostname")
			}

			nameLen := (int(rest[current]) << 8) + int(rest[current+1])
			current += 2
			serverName = string(rest[current : current+nameLen])
		}

		current += extensionDataLength
	}

	if serverName == "" {
		return "", nil, fmt.Errorf("could not find server name")
	}

	readBytes = make([]byte, len(firstByte)+len(versionBytes)+len(restLengthBytes)+len(rest))
	current = 0
	for _, b := range firstByte {
		readBytes[current] = b
		current++
	}

	for _, b := range versionBytes {
		readBytes[current] = b
		current++
	}

	for _, b := range restLengthBytes {
		readBytes[current] = b
		current++
	}

	for _, b := range rest {
		readBytes[current] = b
		current++
	}

	return
}
