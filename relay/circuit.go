package main

import (
	"encoding/binary"
	"net"
)

func link_node(ip_port_exit, group_generator string, max_packet int) (net.Conn, string) {
	exit_conn, err_exit := net.Dial("tcp", ip_port_exit)

	if err_exit != nil {
		return nil, ""
	}

	// Mandamos la maxima capcidad de cada paquete
	max_length_byte := make([]byte, 4)
	binary.LittleEndian.PutUint32(max_length_byte, uint32(max_packet))
	exit_conn.Write(max_length_byte)

	send(exit_conn, []byte(group_generator), max_packet)

	exit_Y_buffer := recv(exit_conn, max_packet) // Recibimos las Y (Dillie-Hellman) calculadas en el relay y en el exit node

	return exit_conn, string(exit_Y_buffer)
}
