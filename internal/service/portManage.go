package service

import (
	"log"
	"strconv"
	"tinycloud/internal/models"
)

func DelInstancePorts(instance models.Instance) {
	err := models.DelPortByInstanceName(instance.Name)
	if err != nil {
		panic(err)
	}
}

func CheckIsPortUsed(param models.InstanceParam) {
	for _, item := range param.PortParams {
		port, err := models.GetInstancePort(item.Protocol, item.Value)
		if err != nil {
			panic(err)
		}
		if port != nil {
			panic("port " + port.Port + " with " + port.Protocol + " protocol is used by " + port.InstanceName)
		}
	}
}

func SavePortUsed(instance models.Instance, param models.InstanceParam) {
	for _, item := range param.PortParams {
		models.AddInstancePort(instance.Id,
			instance.Name,
			instance.AppName,
			item.Protocol,
			item.Value,
		)
	}
}

func getFirstHttpPort(param models.InstanceParam) int {
	for _, item := range param.PortParams {
		if item.Protocol == "http" {
			port, err := strconv.Atoi(item.Value)
			if err != nil {
				log.Println(err)
			}
			return port
		}
	}

	return 0
}
