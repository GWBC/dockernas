package service

import (
	"bufio"
	"dockernas/internal/backend/docker"
	"dockernas/internal/config"
	"dockernas/internal/models"
	"dockernas/internal/utils"
	"encoding/json"
	"io"
	"log"
	"math"
	"os"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
)

func CheckParamIsValid(param models.InstanceParam) {
	match, _ := regexp.MatchString("[a-zA-Z0-9][a-zA-Z0-9_.-]", param.Name)
	if !match {
		panic(param.Name + " not match [a-zA-Z0-9][a-zA-Z0-9_.-] to be a container name")
	}

	for _, item := range append(param.EnvParams, param.OtherParams...) {
		if item.Reg != "" {
			match, err := regexp.MatchString(item.Reg, item.Value)
			if err != nil {
				panic("regexp check faild with " + item.Reg + " " + item.Value + ": " + err.Error())
			}
			if !match {
				panic(item.Value + " is not match " + item.Reg + " on param " + item.Prompt)
			}
		}
	}

	// if param.NetworkMode == models.HOST_MODE {
	// 	if docker.DetectRealSystem() != "linux" {
	// 		panic("host network is only work on linux now")
	// 	}
	// }

	CheckIsPortUsed(param)
}

func AutoAddInstance() {
	containerList := docker.ListContainer()
	for _, container := range containerList {
		name := strings.Split(container.Names[0], "/")
		containerName := name[len(name)-1]

		instance := models.GetInstanceByName(containerName)
		if instance == nil {
			index := strings.Index(container.Image, ":")
			if index < 0 {
				container.Image += ":latest"
			}

			app, template := GetAppByImage(container.Image)
			if app == nil {
				continue
			}

			instance = &models.Instance{}
			instance.AppName = app.Name
			instance.Version = template.Version
			instance.IconUrl = app.IconUrl

			if container.State == "running" {
				instance.State = models.RUNNING
			} else {
				instance.State = models.STOPPED
			}

			instance.Summary = "自动添加"
			instance.Name = containerName
			instance.ContainerID = container.ID
			instance.CreateTime = container.Created * 1000

			//////////////////////////////////////////////////////////////
			instanceParam := models.InstanceParam{}
			instanceParam.Name = instance.Name
			instanceParam.AppName = instance.AppName
			instanceParam.Summary = instance.Summary
			instanceParam.DockerTemplate = *template

			startPort := 0
			endPort := 0

			instanceParam.NetworkMode = models.HOST_MODE

			for _, port := range container.Ports {
				for i, t := range instanceParam.DockerTemplate.PortParams {
					if t.Protocol == "http" || t.Protocol == "ftp" {
						t.Protocol = "tcp"
					}

					if instanceParam.NetworkMode == models.HOST_MODE && port.PublicPort == 0 {
						instanceParam.NetworkMode = models.NOBUND_MODE
					} else if port.IP == "127.0.0.1" {
						instanceParam.NetworkMode = models.LOCAL_MODE
					} else if port.IP == "0.0.0.0" {
						instanceParam.NetworkMode = models.BIRDGE_MODE
					}

					portRange := strings.Split(t.Key, "-")
					if len(portRange) == 2 {
						//范围端口
						if port.Type == t.Protocol {
							start, _ := strconv.Atoi(portRange[0])
							end, _ := strconv.Atoi(portRange[1])
							if port.PrivatePort >= uint16(start) && port.PrivatePort <= uint16(end) {
								if port.PrivatePort == uint16(start) {
									startPort = int(port.PublicPort)
								} else if port.PrivatePort == uint16(end) {
									endPort = int(port.PublicPort)
								} else {
									endPort = int(math.Max(float64(endPort), float64(port.PublicPort)))
								}

								if startPort != 0 && endPort != 0 {
									instanceParam.DockerTemplate.PortParams[i].Value = strconv.Itoa(startPort) + "-" + strconv.Itoa(endPort)
								}

								break
							}
						}
					} else {
						//单一端口
						if port.Type == t.Protocol {
							if t.Key == strconv.Itoa(int(port.PrivatePort)) {
								if port.PublicPort != 0 {
									instanceParam.DockerTemplate.PortParams[i].Value = strconv.Itoa(int(port.PublicPort))
								}

								break
							}
						}
					}
				}
			}

			for _, mount := range container.Mounts {
				for i, volume := range instanceParam.DockerTemplate.LocalVolume {
					if mount.Destination == volume.Key {
						instanceParam.DockerTemplate.LocalVolume[i].Value = mount.Source
						break
					}
				}

				for i, volume := range instanceParam.DockerTemplate.DfsVolume {
					if mount.Destination == volume.Key {
						src := mount.Source
						basePath := config.GetFullDfsPath("/")
						index := strings.Index(src, basePath)
						if index == 0 {
							src = src[len(basePath):]
							src += "/"
						}
						instanceParam.DockerTemplate.DfsVolume[i].Value = src
						break
					}
				}
			}

			strInstanceParam, e := json.Marshal(&instanceParam)
			if e == nil {
				instance.InstanceParamStr = string(strInstanceParam)
			}

			models.AddInstance(instance)
		}
	}
}

func GetInstance() []models.Instance {
	AutoAddInstance()

	instances := models.GetInstance()

	networkInfo := GetNetworkInfo()
	if networkInfo.HttpGatewayEnable {
		for i, instance := range instances {
			proxyConfig := models.GetHttpProxyConfigByInstance(instance.Name)
			if proxyConfig != nil {
				if networkInfo.HttpsEnable {
					instances[i].Url = "https://" + proxyConfig.HostName + "." + networkInfo.Domain
				} else {
					instances[i].Url = "http://" + proxyConfig.HostName + "." + networkInfo.Domain
				}
			}
		}
	}

	return instances
}

func runNewContainer(instance *models.Instance, param models.InstanceParam) {
	var err error

	log.Println("create instance " + instance.Name)
	instance.ContainerID, err = docker.Create(&param)
	instance.InstanceParamStr = utils.GetJsonFromObj(param)

	if err != nil {
		if instance.ContainerID == "" {
			instance.State = models.CREATE_ERROR
		} else {
			instance.State = models.RUN_ERROR
		}
		models.UpdateInstance(instance)
		log.Println(err)
		models.AddEventLog(instance.Id, models.START_EVENT, err.Error())
		panic(err)
	} else {
		models.AddEventLog(instance.Id, models.START_EVENT, "")
	}

	instance.State = models.RUNNING
	models.UpdateInstance(instance)
	SavePortUsed(instance, param)
}

func pullAndRunContainer(instance *models.Instance, param models.InstanceParam, blocking bool) *models.Instance {
	config.GetBasePath() //check base path
	log.Println("pull image " + param.ImageUrl)

	reader, err := docker.PullImage(param.ImageUrl)
	if err != nil {
		instance.State = models.PULL_ERROR
		models.UpdateInstance(instance)
		models.AddEventLog(instance.Id, models.START_EVENT, err.Error())
		panic(err)
	}

	instance.State = models.PULL_IMAGE
	models.UpdateInstance(instance)

	if blocking {
		_, err := io.Copy(log.Writer(), reader)
		if err != nil {
			log.Println(err)
		}
		reader.Close()
		runNewContainer(instance, param)
	} else {
		go func() {
			defer func() {
				err := recover()
				if err != nil {
					log.Println("create instance:", err)
					log.Println(string(debug.Stack()))
				}
				reader.Close()
			}()

			startTime := time.Now().Unix()
			scanner := bufio.NewScanner(reader)
			for scanner.Scan() {
				line := scanner.Text()
				ProcessImagePullMsg(param.ImageUrl, line)
				if (time.Now().Unix() - startTime) >= (60 * 30) { // timeout for 30 minute
					log.Println("pull image " + param.ImageUrl + " time out")
					ReportImagePullStoped(param.ImageUrl)
					instance.State = models.PULL_ERROR
					models.UpdateInstance(instance)
					return
				}
			}
			log.Println("pull image " + param.ImageUrl + " ok")
			ReportImagePullStoped(param.ImageUrl)
			tmp := models.GetInstanceByName(instance.Name)
			if tmp == nil || tmp.Id != instance.Id { //check if instance is deleted
				return
			}
			runNewContainer(instance, param)
		}()
	}

	return instance
}

func GetInstanceByName(name string) models.Instance {
	instance := models.GetInstanceByName(name)

	if instance == nil {
		panic("instance <" + name + "> not exist")
	}

	if instance.State == models.PULL_IMAGE {
		var param models.InstanceParam
		if utils.GetObjFromJson(instance.InstanceParamStr, &param) != nil {
			instance.ImagePullState = GetImagePullState(param.ImageUrl)
		}
	}

	return *instance
}

func CreateInstance(param models.InstanceParam, blocking bool) *models.Instance {
	tInstance := models.GetInstanceByName(param.Name)

	if tInstance != nil {
		panic("instance <" + param.Name + "> exist")
	}

	var instance models.Instance
	instance.Name = param.Name
	instance.Summary = param.Summary
	instance.State = models.PULL_IMAGE
	instance.AppName = param.AppName
	instance.Version = param.Version
	instance.IconUrl = param.IconUrl
	instance.Port = getFirstHttpPort(param)
	instance.InstanceParamStr = utils.GetJsonFromObj(param)
	instance.CreateTime = time.Now().UnixMilli()

	models.AddInstance(&instance)
	pullAndRunContainer(&instance, param, blocking)

	return &instance
}

func EditInstance(instance models.Instance, param models.InstanceParam) {
	DelInstancePorts(instance)
	if docker.IsContainerExist(instance.ContainerID) {
		log.Println("delete comtainer of instance " + instance.Name)
		err := docker.Delete(instance.ContainerID)
		if err != nil {
			models.AddEventLog(instance.Id, models.CONFIG_EVENT, err.Error())
			panic(err)
		} else {
			models.AddEventLog(instance.Id, models.CONFIG_EVENT, "")
		}
	}

	instance.Summary = param.Summary
	instance.State = models.PULL_IMAGE
	instance.AppName = param.AppName
	instance.Version = param.Version
	instance.IconUrl = param.IconUrl
	instance.Port = getFirstHttpPort(param)
	instance.InstanceParamStr = utils.GetJsonFromObj(param)

	models.UpdateInstance(&instance)

	pullAndRunContainer(&instance, param, false)
	models.AddEventLog(instance.Id, models.CONFIG_EVENT, "")
}

func RestartInstance(instance models.Instance) {
	err := docker.Restart(instance.ContainerID)
	if err != nil {
		models.AddEventLog(instance.Id, models.RESTART_EVENT, err.Error())
		panic(err)
	} else {
		models.AddEventLog(instance.Id, models.RESTART_EVENT, "")
	}

	instance.State = models.RUNNING
	models.UpdateInstance(&instance)
}

func StartInstance(instance models.Instance) {
	if instance.ContainerID == "" {
		var param models.InstanceParam
		err := json.Unmarshal([]byte(instance.InstanceParamStr), &param)
		if err != nil {
			log.Println(err)
			panic(err)
		}
		pullAndRunContainer(&instance, param, false)

	} else {
		log.Println("start comtainer of instance " + instance.Name)
		err := docker.Start(instance.ContainerID)
		if err != nil {
			models.AddEventLog(instance.Id, models.START_EVENT, err.Error())
			panic(err)
		}
		instance.State = models.RUNNING
		models.UpdateInstance(&instance)
		models.AddEventLog(instance.Id, models.START_EVENT, "")
	}
}

func StopInstance(instance models.Instance) {
	log.Println("stop comtainer of instance " + instance.Name)
	err := docker.Stop(instance.ContainerID)
	if err != nil {
		models.AddEventLog(instance.Id, models.STOP_EVENT, err.Error())
		panic(err)
	}

	instance.State = models.STOPPED
	models.UpdateInstance(&instance)
	models.AddEventLog(instance.Id, models.STOP_EVENT, "")
}

func DeleteInstance(instance models.Instance) {
	// if instance.State == models.RUNNING {
	// 	StopInstance(instance)
	// }

	DelInstancePorts(instance)
	models.DelInstanceStatData(instance.Name)
	models.DelEvents(instance.Id)

	if docker.IsContainerExist(instance.ContainerID) {
		log.Println("delete container of instance " + instance.Name)
		err := docker.Delete(instance.ContainerID)
		if err != nil {
			models.AddEventLog(instance.Id, models.DELETE_EVENT, err.Error())
			panic(err)
		}
	}
	models.DeleteInstance(&instance)
	err := os.RemoveAll(config.GetAppLocalPath(instance.Name))
	if err != nil {
		log.Println(err)
	}
	// models.AddEventLog(instance.Id, models.DELETE_EVENT, "")
}

func GetInstanceLog(instance models.Instance) string {
	return docker.GetLog(instance.ContainerID)
}
