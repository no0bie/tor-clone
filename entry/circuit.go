// Module that allows the chain connection to happen (circuit creation)

package main

import (
	"encoding/binary"
	"net"
)

func link_node(ip_port, group_generator string, max_packet int) (net.Conn, string) {
	conn, err := net.Dial("tcp", ip_port)

	if err != nil {
		return nil, ""
	}

	// Let the receiving side how many packets we are sending
	max_length_byte := make([]byte, 4)
	binary.LittleEndian.PutUint32(max_length_byte, uint32(max_packet))
	conn.Write(max_length_byte)

	send(conn, []byte(group_generator), max_packet)

	Y_buffer := recv(conn, max_packet)

	return conn, string(Y_buffer)
}
