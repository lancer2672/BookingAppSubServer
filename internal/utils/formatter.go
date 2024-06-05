package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func FormatErrorRes(err error) gin.H {
	return gin.H{"error": err.Error()}
}
func RespondSuccess(ctx *gin.Context, data interface{}) {
	ctx.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"data":    data,
		"message": "success",
	})
}
