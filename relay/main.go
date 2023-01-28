package main

import (
	"encoding/binary"
	"fmt"
	"math/big"
	"net"
	"os"
	"strings"
)

const DIRECTORY_HOST = "http://tor-clone-directory-1:8080/relay"

var max_length int

func process_entry(entry_conn net.Conn) {
	// Recibir el tamaño máximo de paquetes
	max_length_buffer := make([]byte, 4)
	entry_conn.Read(max_length_buffer)
	max_length = int(binary.LittleEndian.Uint32(max_length_buffer))

	gen_group_buffer := recv(entry_conn, max_length)

	gen_group_con := strings.Split(string(gen_group_buffer), ";")

	g := gen_group_con[0] // generador
	G := gen_group_con[1] // grupo

	group_generator := g + ";" + G // concatenacion de generador, grupo y informarcion del exit node que se manda al relay

	exit_conn, exit_Y := link_node(gen_group_con[2], group_generator, max_length) // Empezamos la cadena de conexiones y recibimos ambos Dillie-Hellman calculados en los nodos

	if exit_conn == nil {
		entry_conn.Write([]byte("Impossible to reach exit node"))
		entry_conn.Close()
		return
	}

	Y, b, G_big := calc_dh(g, G) // Calculamos Dillie-Hellman TODO

	all_Y := exit_Y + ";" + Y.String()

	send(entry_conn, []byte(all_Y), max_length) // Mandamos todos los dillie-hellman al cliente para que este pueda codificar

	X_buffer := recv(entry_conn, max_length)

	X := new(big.Int)
	X.SetString(string(X_buffer), 10) // Dillie-hellam calculado en el cliente

	send(exit_conn, X_buffer, max_length) // Lo mandamos al resto de nodos

	secret := new(big.Int)
	secret.Exp(X, &b, &G_big) // Calculamos el secreto

	MSG_buffer := recv(entry_conn, max_length)

	decrypted_layer_2 := decrypt(*secret, MSG_buffer)

	fmt.Println("=============================================")
	fmt.Println("Decrypting data with layer 2 secret: ")
	fmt.Println(decrypted_layer_2)

	send(exit_conn, []byte(decrypted_layer_2), max_length)

	encrypted_layer_3 := recv(exit_conn, max_length)
	encrypted_layer_2_3 := encrypt(*secret, encrypted_layer_3)

	fmt.Println("============================================")
	fmt.Println("Encrypting layer 3 data with layer 2 secret: ")
	fmt.Println(encrypted_layer_2_3)

	send(entry_conn, encrypted_layer_2_3, max_length)
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
	fmt.Println("#     Relay Node Started     #")
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

		go process_entry(connection)
	}
}
