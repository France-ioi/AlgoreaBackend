package app

import "github.com/gin-gonic/gin"

// InitApplication starts the application and return the router as handler
func InitApplication() *gin.Engine {
	return InitRouter()
}
