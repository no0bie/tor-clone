package main

import (
	"encoding/binary"
	"fmt"
	"math/big"
	"net"
	"os"
	"strings"
)

const DIRECTORY_HOST = "http://tor-clone-directory-1:8080/entry"

var max_length int

func process_client(client_conn net.Conn) {
	// Get max size of packets that we can send and receive
	max_length_buffer := make([]byte, 4)
	client_conn.Read(max_length_buffer)
	max_length = int(binary.LittleEndian.Uint32(max_length_buffer))

	gen_group_buffer := recv(client_conn, max_length)

	gen_group_con := strings.Split(string(gen_group_buffer), ";")

	g := gen_group_con[0] // Generator
	G := gen_group_con[1] // Group

	relay_ip_port := gen_group_con[2] // ip:port of relay node that we want to connect to
	exit_ip_port := gen_group_con[3]  // ip:port of the exit node, we want the relay to connect to this node, so we forward this on along the generator and group

	group_generator_relay_node := g + ";" + G + ";" + exit_ip_port + ";" // Concatenate generator, group and connection info of the exit node

	relay_conn, relay_Y_Exit_Y := link_node(relay_ip_port, group_generator_relay_node, max_length) // Start chain connection and receive Dillie-Hellman from relay and exit

	if relay_conn == nil {
		client_conn.Write([]byte{})
		client_conn.Close()
		return
	}

	Y, b, G_big := calc_dh(g, G) // Calculate Diffie-Hellman

	all_Y := relay_Y_Exit_Y + ";" + Y.String()

	send(client_conn, []byte(all_Y), max_length) // Send all Diffie-Hellman to the client

	X_buffer := recv(client_conn, max_length)

	X := new(big.Int)
	X.SetString(string(X_buffer), 10) // Client calculated Diffie-Hellmanm

	send(relay_conn, X_buffer, max_length) // Forwarding the client Diffie-Hellmanm

	secret := new(big.Int)
	secret.Exp(X, &b, &G_big) // Secret calculation

	MSG_buffer := recv(client_conn, max_length)

	decrypted_layer_1 := decrypt(*secret, MSG_buffer)

	fmt.Println("=============================================")
	fmt.Println("Decrypting data with layer 1 secret: ")
	fmt.Println(decrypted_layer_1)

	send(relay_conn, decrypted_layer_1, max_length)

	encrypted_layer_2_3 := recv(relay_conn, max_length)
	encrypted_layer_1_2_3 := encrypt(*secret, encrypted_layer_2_3)

	fmt.Println("===================================================")
	fmt.Println("Encrypting layer 2 and 3 data with layer 1 secret: ")
	fmt.Println(encrypted_layer_1_2_3)

	send(client_conn, encrypted_layer_1_2_3, max_length)

}

func main() {
	server_port := "1234"

	if sp := os.Getenv("PORT"); sp != "" {
		server_port = sp
	}

	server, err := net.Listen("tcp", "0.0.0.0:"+server_port)

	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}

	defer server.Close()

	fmt.Println("##############################")
	fmt.Println("#                            #")
	fmt.Println("#     Entry Node Started     #")
	fmt.Println("#                            #")
	fmt.Println("##############################")

	post_ip_port(server_port, DIRECTORY_HOST)

	for {
		connection, err := server.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		fmt.Println("\n > New Client:")

		go process_client(connection)
	}
}
