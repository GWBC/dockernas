package models

import (
	"dockernas/internal/config"
	"errors"
	"log"
	"net/url"

	"gorm.io/gorm"
)

const (
	PULL_IMAGE   = 0
	CREATE_ERROR = 1
	RUN_ERROR    = 2
	RUNNING      = 3
	STOPPED      = 4
	PULL_ERROR   = 5
)

type Instance struct {
	Id               int    `json:"id"  gorm:"primary_key;auto_increment"`
	ContainerID      string `json:"containerID"`
	Summary          string `json:"summary"`
	State            int    `json:"state"`
	IconUrl          string `json:"iconUrl"`
	Port             int    `json:"port"`
	Url              string `json:"url"`
	Name             string `json:"name"  gorm:"unique;not null"`
	AppName          string `json:"appName"`
	Version          string `json:"version"`
	InstanceParamStr string `json:"instanceParamStr" gorm:"type:varchar(1024)"` //store json str
	CreateTime       int64  `json:"createTime"`
	ImagePullState   string `json:"imagePullState"`
	DockerSvrIP      string `json:"dockerSvrIP" gorm:"-"`
}

func GetDockerSvrIP() string {
	ip := config.GetConfig("dockerSvrIP", "")
	if len(ip) != 0 {
		ip = "tcp://" + ip
		urlObj, e := url.Parse(ip)
		if e != nil {
			ip = ""
		} else {
			ip = urlObj.Hostname()
		}
	}

	return ip
}

func AddInstance(instance *Instance) {
	err := GetDb().Create(instance).Error
	if err != nil {
		log.Println(err)
		panic(err)
	}
}

func UpdateInstance(instance *Instance) {
	err := GetDb().Model(&Instance{}).Where("id = ?", instance.Id).Save(instance).Error
	if err != nil {
		log.Println(err)
		panic(err)
	}
}

func DeleteInstance(instance *Instance) {
	err := GetDb().Where("name = ?", instance.Name).Delete(instance).Error
	if err != nil {
		log.Println(err)
		panic(err)
	}
}

func GetInstance() []Instance {
	var instances []Instance
	err := GetDb().Find(&instances).Error
	if err != nil {
		log.Println(err)
		panic(err)
	}

	for i := 0; i < len(instances); i++ {
		instances[i].DockerSvrIP = GetDockerSvrIP()
	}

	return instances
}

func GetInstanceByName(name string) *Instance {
	var instance Instance
	err := GetDb().First(&instance, "name=?", name).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		log.Println(err)
		panic(err)
	}

	instance.DockerSvrIP = GetDockerSvrIP()

	return &instance
}

func GetInstanceByID(id string) *Instance {
	var instance Instance
	err := GetDb().First(&instance, "container_id=?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		log.Println(err)
		panic(err)
	}

	instance.DockerSvrIP = GetDockerSvrIP()

	return &instance
}
