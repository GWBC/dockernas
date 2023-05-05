package service

import (
	"dockernas/internal/config"
	"dockernas/internal/models"
	"dockernas/internal/utils"
	"log"
	"strconv"
	"strings"
)

func DelInstancePorts(instance models.Instance) {
	err := models.DelPortByInstanceName(instance.Name)
	if err != nil {
		panic(err)
	}
}

func CheckIsPortUsed(param models.InstanceParam) {
	for _, item := range param.PortParams {
		if item.Value == "" {
			continue
		}

		port, err := models.GetInstancePort(item.Protocol, item.Value)
		if err != nil {
			panic(err)
		}

		if port != nil {
			if port.InstanceName == param.Name {
				continue
			} else {
				panic("port " + port.Port + " with " + port.Protocol + " protocol is used by " + port.InstanceName)
			}
		}

		portRange := strings.Split(item.Value, "-")
		if len(portRange) == 0 || len(portRange) > 2 {
			panic("port or start - end")
		}

		startPort, err := strconv.Atoi(portRange[0])
		if err != nil {
			panic(portRange[0] + " is not a valide port value")
		}

		if startPort >= 65535 {
			panic(portRange[0] + " is greater than max port value 65535")
		}

		endPort := startPort
		if len(portRange) == 2 {
			endPort, err = strconv.Atoi(portRange[1])
			if err != nil {
				panic(portRange[1] + " is not a valide port value")
			}

			if endPort >= 65535 {
				panic(portRange[1] + " is greater than max port value 65535")
			}
		}

		for port := startPort; port <= endPort; port++ {
			if utils.IsPortUsed(getCheckHost(), item.Protocol, strconv.Itoa(port)) {
				panic("port " + strconv.Itoa(port) + " with " + item.Protocol + " protocol is used")
			}
		}
	}
}

func getCheckHost() string {
	host := models.GetDockerSvrIP()
	if len(host) == 0 {
		host = "localhost"

		if config.IsRunInConainer() {
			host = "host.docker.internal"
		}
	}

	return host
}

func SavePortUsed(instance *models.Instance, param models.InstanceParam) {
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
