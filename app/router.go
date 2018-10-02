package app

import (
	"github.com/gin-gonic/gin"
)

// InitRouter creates, inits and returns the GIN router
func InitRouter() *gin.Engine {

	r := gin.New()

	return r
}
