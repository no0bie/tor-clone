package main

import (
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// Entry point for docker to check if the directory is alive
func alive(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, map[string]bool{"alive": true})
}

func main() {

	// Initiate the udp server to let the nodes know what their connecting IP is
	server, err := net.ListenPacket("udp", "0.0.0.0:1235")

	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}

	fmt.Println("UDP started on port 1235")

	defer server.Close()

	// Start directory server
	router := gin.Default()

	router.GET("/", alive)

	router.GET("/entry", getEntry)
	router.POST("/entry", postEntry)

	router.GET("/relay", getRelay)
	router.POST("/relay", postRelay)

	router.GET("/exit", getExit)
	router.POST("/exit", postExit)

	router.Run("0.0.0.0:8080")

	fmt.Println("Directory started on port 8080")
}
