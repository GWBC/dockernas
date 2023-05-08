package server

import (
	"dockernas/internal/config"
	"dockernas/internal/middleware"
	"dockernas/internal/models"
	"log"
	"strings"

	"github.com/gin-gonic/gin"
)

func StartServer() {
	models.AutoInsertLocalhost()

	gin.DisableConsoleColor()
	gin.DefaultWriter = log.Writer()
	gin.DefaultErrorWriter = log.Writer()

	router := gin.New()
	router.Use(
		gin.Logger(),
		middleware.Recovery(),
	)

	initStaticFileRouter(router)
	registerRoutes(router)
	router.NoRoute(func(ctx *gin.Context) {
		if strings.Index(ctx.Request.URL.Path, "/index/") == 0 ||
			ctx.Request.URL.Path == "/login" ||
			ctx.Request.URL.Path == "/basepath" {
			ctx.Request.URL.Path = "/"
			router.HandleContext(ctx)
		}
	})

	router.Run(config.GetBindAddr())
}
