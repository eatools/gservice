package application

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

var (
	Version = ""
	Git     = ""
	Build   = ""
)

func Status(c *gin.Context) {
	c.Data(http.StatusOK, "application/text", []byte(fmt.Sprintf("VERSION=%v \nGIT=%v \nBUILD TIME=%v \n", Version, Git, Build)))
}
