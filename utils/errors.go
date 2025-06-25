package utils

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

var (
	SUCCESS = "0X200"
)

type Error struct {
	ErrorCode    string      `json:"error_code"`
	ErrorMessage string      `json:"error_message"`
	Data         interface{} `json:"data,omitempty"`
}

func GetError(code string, message string, data interface{}) Error {
	return Error{
		ErrorCode:    code,
		ErrorMessage: message,
		Data:         data,
	}
}

func GenerateReturnData(c *gin.Context, code string, message string, data interface{}) {
	c.JSON(http.StatusOK, GetError(code, message, data))
}
