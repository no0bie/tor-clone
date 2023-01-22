package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type entry_node_struct struct {
	Conn string `json:"conn"`
}

var entry_nodes = []entry_node_struct{}

func getEntry(c *gin.Context) {
	rand.Seed(time.Now().Unix())
	if len(entry_nodes) > 0 {
		c.IndentedJSON(http.StatusOK, entry_nodes[rand.Intn(len(entry_nodes))])
	} else {
		c.IndentedJSON(http.StatusOK, entry_node_struct{""})
	}
}

func postEntry(c *gin.Context) {
	var new_entry_node entry_node_struct

	if err := c.BindJSON(&new_entry_node); err != nil {
		return
	}

	entry_nodes = append(entry_nodes, new_entry_node)

	fmt.Println(new_entry_node)

	c.IndentedJSON(http.StatusCreated, new_entry_node)
}
