package main

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"math/big"
	"net"
	"net/http"
	"os"
	"strings"
)

const (
	DIRECTORY_HOST = "http://tor-clone-directory-1:8080/exit"
)

var max_length int

func process_entry(relay_conn net.Conn) {
	// Get max size of packets that we can send and receive
	max_length_buffer := make([]byte, 4)
	relay_conn.Read(max_length_buffer)
	max_length = int(binary.LittleEndian.Uint32(max_length_buffer))

	gen_group_buffer := recv(relay_conn, max_length)

	gen_group_con := strings.Split(string(gen_group_buffer), ";")

	g := gen_group_con[0] // Generator
	G := gen_group_con[1] // Group

	Y, b, G_big := calc_dh(g, G) // Calculate Diffie-Hellman

	send(relay_conn, []byte(Y.String()), max_length) // Send all Diffie-Hellman to the client

	X_buffer := recv(relay_conn, max_length)

	X := new(big.Int)
	X.SetString(string(X_buffer), 10) // Client calculated Diffie-Hellmanm

	secret := new(big.Int)
	secret.Exp(X, &b, &G_big) // Secret calculation

	MSG_buffer := recv(relay_conn, max_length)

	decrypted_layer_3 := decrypt(*secret, MSG_buffer)

	msg := string(decrypted_layer_3)

	new_msg := []byte{}

	// match, _ := regexp.MatchString("^[http://|https://].*$", msg) -> Pointer failing after few requests, decided to go for a simpler approach

	if strings.HasPrefix(msg, "http://") || strings.HasPrefix(msg, "https://") {
		entry_node, _ := http.Get(msg)
		new_msg, _ = ioutil.ReadAll(entry_node.Body)
	} else {
		new_msg = []byte("You sent '" + msg + "' which does not try to access any external services.<br />This message comes from the selected exit node (Layer 3)")
	}

	encrypted_layer_3 := encrypt(*secret, new_msg)

	fmt.Println("=============================================")
	fmt.Println("Encrypting reponse data with layer 3 secret: ")
	fmt.Println(encrypted_layer_3)

	send(relay_conn, encrypted_layer_3, max_length)
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

	fmt.Println("###############################")
	fmt.Println("#                             #")
	fmt.Println("#      Exit Node Started      #")
	fmt.Println("#                             #")
	fmt.Println("###############################")

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
