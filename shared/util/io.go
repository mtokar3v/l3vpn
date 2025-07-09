package util

import (
	"io"
	"net"
)

const bufSize = 255

func WritePacket(conn *net.TCPConn, data []byte) (int, error) {
	n := len(data)

	for len(data) > 0 {
		chunk := data
		if len(chunk) > bufSize {
			chunk = data[:bufSize]
		}
		if _, err := conn.Write([]byte{byte(len(chunk))}); err != nil {
			return n - len(data), err
		}
		if _, err := conn.Write(chunk); err != nil {
			return n - len(data), err
		}
		data = data[len(chunk):]
	}
	// write null terminator
	_, err := conn.Write([]byte{0})
	return n, err
}

func ReadPacket(conn *net.TCPConn) ([]byte, error) {
	var res []byte
	bufLen := make([]byte, 1)

	for {
		if _, err := io.ReadFull(conn, bufLen); err != nil {
			return nil, err
		}
		n := int(bufLen[0])
		// null terminator check
		if n == 0 {
			break
		}
		buf := make([]byte, n)
		if _, err := io.ReadFull(conn, buf); err != nil {
			return nil, err
		}
		res = append(res, buf...)
	}

	return res, nil
}
