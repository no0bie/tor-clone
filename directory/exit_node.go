package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type exit_node_struct struct {
	Conn string `json:"conn"`
}

var exit_nodes = []exit_node_struct{}

func getExit(c *gin.Context) {
	rand.Seed(time.Now().Unix())
	if len(exit_nodes) > 0 {
		c.IndentedJSON(http.StatusOK, exit_nodes[rand.Intn(len(exit_nodes))])
	} else {
		c.IndentedJSON(http.StatusOK, exit_node_struct{""})
	}
}

func postExit(c *gin.Context) {
	var new_exit_node exit_node_struct

	if err := c.BindJSON(&new_exit_node); err != nil {
		return
	}

	exit_nodes = append(exit_nodes, new_exit_node)

	fmt.Println(new_exit_node)
	c.IndentedJSON(http.StatusCreated, new_exit_node)
}
