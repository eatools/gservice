package application

import (
	"net/http"

	"gopkg.in/ini.v1"

	"github.com/gin-gonic/gin"
)

func NewServer(conf *ini.File) *gin.Engine {

	httpServer := gin.Default()

	PProfRegister(httpServer) // 性能
	httpServer.GET("/status", Certificate)

	return httpServer
	//return httpServer.Run(conf.HttpAddress)
}

func Certificate(c *gin.Context) {
	c.Data(http.StatusOK, "application/text", []byte("success"))
}
