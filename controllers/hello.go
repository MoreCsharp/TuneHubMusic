package controllers

import (
	"github.com/gin-gonic/gin"
)

func Hello(c *gin.Context) {
	c.JSON(200, gin.H{
		"code":    0,
		"message": "Hello, Gin!",
		"data":    nil,
	})
}
