package api

import (
	"dockernas/internal/models"
	"dockernas/internal/service"

	"github.com/gin-gonic/gin"
)

func AddDockerSvrInfo(c *gin.Context) {
	var param models.DockerSvrInfo
	c.BindJSON(&param)

	service.AddDockerSvrInfo(param)

	c.JSON(200, gin.H{
		"msg": "ok",
	})
}

func UpdateDockerSvrInfo(c *gin.Context) {
	var param models.DockerSvrInfo
	c.BindJSON(&param)

	service.UpdateDockerSvrInfo(param)

	c.JSON(200, gin.H{
		"msg": "ok",
	})
}

func DeleteDockerSvrInfo(c *gin.Context) {
	var param int
	c.BindJSON(&param)

	service.DeleteDockerSvrInfo(param)

	c.JSON(200, gin.H{
		"msg": "ok",
	})
}

func GetDockerSvrInfos(c *gin.Context) {
	infos := service.GetDockerSvrInfos()

	c.JSON(200, gin.H{
		"list": infos,
	})
}
