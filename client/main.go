package main

import (
	"bytes"
	"crypto/rand"
	b64 "encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"text/template"
)

const (
	DIRECTORY_HOST         = "http://tor-clone-directory-1:8080"  // Directory host to get all node addressess
	SAMPLE_END_SERVER_HOST = "http://tor-clone-end-server-1:8080" // The end server that will tell us if our circuit is working
	MAX_PACKET             = 2                                    // Max packet size that all nodes will receive and agree on
	GENERATOR              = 2                                    // g = 69
	GROUP                  = "179769313486231590770839156793787453197860296048756011706444423684197180216158519368947833795864925541502180565485980503646440548199239100050792877003355816639229553136239076508735759914822574862575007425302077447712589550957937778424442426617334727629299387668709205606050270810842907692932019128194467627007"
)

type node_struct struct {
	Conn string `json:"conn"`
}

type response struct {
	ClientIP string
	Action   string
	Circuit  string
	Response string
}

// Get our IP so we can display it on the webpage
func get_client_ip() string {
	conn, err := net.Dial("udp", "tor-clone-directory-1:1235")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String()
}

// Helper function to reverse bigInt arrays
func reverse(arr []big.Int) []big.Int {
	reversed := []big.Int{}

	for i := len(arr) - 1; i >= 0; i -= 1 {
		reversed = append(reversed, arr[i])
	}

	return reversed
}

// Diffie-Hellman
func calc_dh(recv []string) ([]big.Int, big.Int) {
	secrets := []big.Int{}

	max_min, min := big.NewInt(10000-3), big.NewInt(3)

	big_generator := big.NewInt(GENERATOR)
	big_group := new(big.Int)
	big_group, _ = big_group.SetString(GROUP, 10)

	a, _ := rand.Int(rand.Reader, max_min)
	a.Add(a, min)

	X := new(big.Int)
	X.Exp(big_generator, a, big_group) // (g^a) mod G

	for _, recv_val := range recv {
		recv_val_big := new(big.Int)
		recv_val_big.SetString(recv_val, 10)

		secret := new(big.Int)
		secret.Exp(recv_val_big, a, big_group)

		secrets = append(secrets, *secret)
	}

	return secrets, *X
}

// Encrypt the query with AES-GCM (Galois/Counter Mode)
func encrypt_all(secrets []big.Int, msg string) []byte {
	msg_bytes := []byte(msg)

	for _, secret := range secrets {
		msg_bytes = encrypt(secret, msg_bytes)
	}
	return msg_bytes
}

// Decrypt the response with AES-GCM (Galois/Counter Mode)
func decrypt_all(secrets []big.Int, encrypted []byte) string {
	for _, secret := range secrets {
		encrypted = decrypt(secret, encrypted)

	}

	return string(encrypted)
}

// Divides the data into our max_length that has been agreed upon by client and nodes
func split_bytes(data []byte) [][]byte {
	splits := [][]byte{}

	l, r := 0, MAX_PACKET
	for ; r < len(data); l, r = r, r+MAX_PACKET {
		splits = append(splits, data[l:r])
	}

	splits = append(splits, data[l:])

	return splits
}

func recv(conn net.Conn) []byte {
	// How many packets will we receive?
	values_buffer := make([]byte, 4)
	conn.Read(values_buffer)
	n_packet := int(binary.LittleEndian.Uint32(values_buffer))

	values := []byte{}

	// Receive all packets
	for i := 0; i < n_packet; i += 1 {
		values_buffer = make([]byte, MAX_PACKET)
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

func send(conn net.Conn, value []byte) {
	bs64 := b64.StdEncoding.EncodeToString(value)
	value_split := split_bytes([]byte(bs64))

	value_split_bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(value_split_bytes, uint32(len(value_split)))

	// Let the receiving side how many packets we are sending
	conn.Write(value_split_bytes)

	for _, bytes := range value_split {
		conn.Write(bytes) // Send them all
	}
}

// Parse GET requests
func read_get_req(req http.Response) node_struct {
	body, _ := ioutil.ReadAll(req.Body)
	var struct_req node_struct
	json.Unmarshal(body, &struct_req)

	return struct_req
}

// Perform GET request on the directory to get all nodes of the circuit
func create_circuit() (string, string, string) {
	entry_node, _ := http.Get(DIRECTORY_HOST + "/entry")
	entry_node_info := read_get_req(*entry_node)

	relay_node, _ := http.Get(DIRECTORY_HOST + "/relay")
	relay_node_info := read_get_req(*relay_node)

	exit_node, _ := http.Get(DIRECTORY_HOST + "/exit")
	exit_node_info := read_get_req(*exit_node)

	return entry_node_info.Conn, relay_node_info.Conn, exit_node_info.Conn
}

func connection_handler(entry, relay_exit, msg string) string {

	// Establish connection
	connection, err := net.Dial("tcp", entry)

	if err != nil {
		panic(err)
	}

	// Send max size of packets that will be sent
	max_length := make([]byte, 4)
	binary.LittleEndian.PutUint32(max_length, uint32(MAX_PACKET))
	connection.Write(max_length)

	// Send generator and group to all nodes, also send the circuit information (the connection address and port of each node)
	send(connection, []byte(strconv.Itoa(GENERATOR)+";"+GROUP+relay_exit))

	values := recv(connection)

	recv_values := strings.Split(string(values), ";")

	encryption_secrets, to_send := calc_dh(recv_values)

	send(connection, []byte(to_send.String()))

	encrypted_data := encrypt_all(encryption_secrets, msg)
	encrypted_dat_tmp := encrypted_data
	fmt.Println("\n > New Query:")

	fmt.Println("=========== Raw Query ===========")
	fmt.Println(" - Query: " + msg)

	fmt.Println("======== Encrypted Query ========")
	fmt.Println(" - String -> " + string(encrypted_dat_tmp))
	fmt.Print("\n - Bytes -> ")
	fmt.Println(encrypted_dat_tmp)

	send(connection, encrypted_data)

	encrypted_layer_1_2_3 := recv(connection)

	response := decrypt_all(reverse(encryption_secrets), encrypted_layer_1_2_3)
	fmt.Println("========= Decoded Response =========")
	fmt.Println(response)

	connection.Close()

	return response
}

func main() {
	cur_ip := get_client_ip()
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	tmpl := template.Must(template.ParseFiles("layout.html"))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		if r.URL.Path != "/" {
			fmt.Fprint(w, "404 not found")
			return
		}

		tmpl_data := response{
			Circuit:  "",
			Response: "",
			Action:   "",
			ClientIP: cur_ip,
		}

		if r.Method == "POST" {
			if err := r.ParseForm(); err != nil {
				fmt.Fprintf(w, "ParseForm() err: %v", err)
				return
			} else {
				entry, relay, exit := create_circuit()

				relay_exit_nodes := ";" + relay + ";" + exit + ";"

				circuit := "Client (you) ⇔ " + entry + " (Layer 1) ⇔ " + relay + " (Layer 2) ⇔ " + exit + " (Layer 3) ⇔ End Server"

				msg := r.FormValue("action")

				switch r.FormValue("type") {
				case "search":
					tmpl_data.Response = connection_handler(entry, relay_exit_nodes, "https://www.google.com/search?q="+strings.ReplaceAll(msg, " ", "%20"))
				case "end":
					tmpl_data.Response = connection_handler(entry, relay_exit_nodes, SAMPLE_END_SERVER_HOST+"?msg="+strings.ReplaceAll(msg, " ", "%20"))
				default:
					tmpl_data.Response = connection_handler(entry, relay_exit_nodes, msg)
				}

				tmpl_data.Circuit = circuit
				tmpl_data.Action = msg

				os.WriteFile("./static/index.html", []byte(tmpl_data.Response), 0644)
			}
		}

		tmpl.Execute(w, tmpl_data)
	})

	fmt.Println("Client started at port 8080")

	// Set up out front end client
	if err := http.ListenAndServe("0.0.0.0:8080", nil); err != nil {
		panic(err)
	}
}
