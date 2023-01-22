// Module to post IP of node in the directory

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
)

// We connect to the directory to see what is our IP
func get_ip() string {
	conn, err := net.Dial("udp", "tor-clone-directory-1:1235")
	if err != nil {
		return ""
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String()
}

func post_ip_port(port, directory string) {
	dat, _ := os.ReadFile("/etc/hostname") // Fail safe in case get_ip() fails and we only send our hostname instead of IP

	ip_raw := get_ip()

	dat_string := strings.TrimRight(string(dat), "\n")

	if ip_raw != "" {
		dat_string = ip_raw
	}

	payload := map[string]string{"conn": dat_string + ":" + port}

	json_payload, _ := json.Marshal(payload)

	a, err := http.Post(directory, "application/json", bytes.NewBuffer(json_payload))

	fmt.Println(a)
	fmt.Println(err)
}
