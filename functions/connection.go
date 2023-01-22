// Module that handles all of the connection logistics

package main

import (
	"bytes"
	b64 "encoding/base64"
	"encoding/binary"
	"fmt"
	"net"
)

// Divides the data into our max_length that has been agreed upon by client and nodes
func split_bytes(data []byte, max_length int) [][]byte {
	splits := [][]byte{}

	l, r := 0, max_length
	for ; r < len(data); l, r = r, r+max_length {
		splits = append(splits, data[l:r])
	}

	splits = append(splits, data[l:])
	return splits
}

func recv(conn net.Conn, max_packet_length int) []byte {
	// How many packets will we receive?
	values_buffer := make([]byte, 4)
	conn.Read(values_buffer)
	n_packet := int(binary.LittleEndian.Uint32(values_buffer))

	values := []byte{}

	// Receive all packets
	for i := 0; i < n_packet; i += 1 {
		values_buffer = make([]byte, max_packet_length)
		conn.Read(values_buffer)
		values = append(values, values_buffer...)
	}

	values = bytes.TrimRight(values, "\x00")                    // Trim padding bytes
	values, err := b64.StdEncoding.DecodeString(string(values)) // Decode from b64 to bytes

	if err != nil {
		fmt.Println(err)
	}

	return values
}

func send(conn net.Conn, value []byte, max_packet_length int) {
	bs64 := b64.StdEncoding.EncodeToString(value)
	value_split := split_bytes([]byte(bs64), max_packet_length)

	value_split_bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(value_split_bytes, uint32(len(value_split)))

	// Let the receiving side how many packets we are sending
	conn.Write(value_split_bytes)

	for _, bytes := range value_split {
		conn.Write(bytes) // Send them all
	}
}
