package resp

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type H = map[string]interface{}

func Error(c *gin.Context, message string) {
	c.AbortWithStatusJSON(http.StatusBadRequest, H{"error": message})
}

func ErrorWithCode(c *gin.Context, code int, message string) {
	c.AbortWithStatusJSON(code, H{"error": message})
}

func JSON(c *gin.Context, code int, data interface{}) {
	c.JSON(code, data)
}

func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, H{
		"success": true,
		"data":    data,
	})
}
