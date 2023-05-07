package service

import (
	"dockernas/internal/models"
)

func AddDockerSvrInfo(info models.DockerSvrInfo) {
	models.AddDockerSvrInfo(&info)
}

func UpdateDockerSvrInfo(info models.DockerSvrInfo) {
	models.UpdateDockerSvrInfo(&info)
}

func DeleteDockerSvrInfo(id int) {
	models.DeleteDockerSvrInfo(id)
}

func GetDockerSvrInfos() []models.DockerSvrInfo {
	return models.GetDockerSvrInfos()
}
