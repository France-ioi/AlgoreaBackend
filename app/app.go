package app

import (
	"os"

	"github.com/gin-gonic/gin"
)

// InitApplication starts the application and return the router as handler
func InitApplication() *gin.Engine {

	// load config + handle errors
	if err := Config.Load(); err != nil {
		os.Exit(1)
	}

	return InitRouter()
}
