// A simple http server that will tell you what is your IP

package main

import (
	"fmt"
	"html/template"
	"net"
	"net/http"
	"strings"
)

type parse struct {
	Ip  string
	Msg string
}

func get_ip(r *http.Request) string {
	//Get IP from the X-REAL-IP header
	ip := r.Header.Get("X-REAL-IP")
	netIP := net.ParseIP(ip)
	if netIP != nil {
		return ip
	}

	//Get IP from X-FORWARDED-FOR header
	ips := r.Header.Get("X-FORWARDED-FOR")
	splitIps := strings.Split(ips, ",")
	for _, ip := range splitIps {
		netIP := net.ParseIP(ip)
		if netIP != nil {
			return ip
		}
	}

	//Get IP from RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return err.Error()
	}
	netIP = net.ParseIP(ip)
	if netIP != nil {
		return ip
	}

	return "No valid IP found"
}

func main() {
	tmpl := template.Must(template.ParseFiles("layout.html"))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		if r.URL.Path != "/" {
			fmt.Fprint(w, "404 not found")
			return
		}

		ip := get_ip(r)
		msg := r.URL.Query().Get("msg")

		if msg == "" {
			msg = "(no message was sent)"
		}

		tmpl_data := parse{Ip: ip, Msg: msg}

		tmpl.Execute(w, tmpl_data)
	})

	http.ListenAndServe(":8080", nil)
}
