package server

import (
	"tinycloud/internal/api"
	"tinycloud/internal/middleware"

	"github.com/gin-gonic/gin"
)

func registerRoutes(router *gin.Engine) {
	apiv1 := router.Group("/api", middleware.Authentication())
	{
		apiv1.GET("app", api.GetApps)
		apiv1.GET("app/:name", api.GetAppByName)

		apiv1.POST("instance", api.PostInstance)
		apiv1.GET("instance", api.GetInstance)
		apiv1.GET("instance/:name", api.GetInstanceByName)
		apiv1.PATCH("instance/:name", api.PatchInstance)
		apiv1.DELETE("instance/:name", api.DeleteInstance)

		apiv1.GET("instance/:name/log", api.GetInstanceLog)
		apiv1.GET("instance/:name/event", api.GetInstanceEvent)

		apiv1.GET("filesystem", api.GetDfsDir)

		apiv1.POST("login", api.Login)
	}
}
