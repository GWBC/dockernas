package server

import (
	"strings"
	"tinycloud/internal/middleware"

	"github.com/gin-gonic/gin"
)

func StartServer() {
	router := gin.New()
	router.Use(
		gin.Logger(),
		middleware.Recovery(),
	)

	router.StaticFile("/", "./frontend/dist/index.html")
	router.StaticFile("/favicon.ico", "./frontend/dist/favicon.ico")
	router.Static("/assets", "./frontend/dist/assets")
	router.Static("/apps", "./apps")
	router.NoRoute(func(ctx *gin.Context) {
		if strings.Index(ctx.Request.URL.Path, "/index/") == 0 {
			ctx.Request.URL.Path = "/"
			router.HandleContext(ctx)
		}
	})

	registerRoutes(router)
	router.Run()
}
