package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type relay_node_struct struct {
	Conn string `json:"conn"`
}

var relay_nodes = []relay_node_struct{}

func getRelay(c *gin.Context) {
	rand.Seed(time.Now().Unix())
	if len(relay_nodes) > 0 {
		c.IndentedJSON(http.StatusOK, relay_nodes[rand.Intn(len(relay_nodes))])
	} else {
		c.IndentedJSON(http.StatusOK, relay_node_struct{""})
	}
}

func postRelay(c *gin.Context) {
	var new_relay_node relay_node_struct

	if err := c.BindJSON(&new_relay_node); err != nil {
		return
	}

	relay_nodes = append(relay_nodes, new_relay_node)
	fmt.Println(new_relay_node)
	c.IndentedJSON(http.StatusCreated, new_relay_node)
}
