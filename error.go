package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type HTTPErrorResp struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

func respError(c *gin.Context, status int, msg string) {
	c.JSON(status, HTTPErrorResp{status, msg})
}

func ErrorHandler(logger *Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			respError(c, http.StatusInternalServerError, "Internal server error")
		}

		for _, err := range c.Errors {
			logger.Errorf("%s", err)
		}
	}
}
