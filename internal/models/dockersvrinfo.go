package models

import "log"

type DockerSvrInfo struct {
	Id   int    `json:"id"  gorm:"primary_key;auto_increment"`
	Name string `json:"name" gorm:"uniqueIndex"`
	IP   string `json:"ip" gorm:"uniqueIndex"`
	Use  int    `json:"use"`
}

const (
	DockerSvrNotUse = iota
	DockerSvrUse
)

func AutoInsertLocalhost() {
	info := DockerSvrInfo{}
	info.Id = 0
	info.Name = "本机"
	info.IP = "localhost"

	useInfo := GetUseSvrInfo()
	if useInfo == nil {
		info.Use = DockerSvrUse
	} else {
		info.Use = DockerSvrNotUse
	}

	err := GetDb().Exec("replace into t_docker_svr_info(id, name, ip, use) values(?, ?, ?, ?)", info.Id, info.Name, info.IP, info.Use).Error
	if err != nil {
		log.Println(err)
		panic(err)
	}
}

func AddDockerSvrInfo(info *DockerSvrInfo) {
	var err error

	db := GetDb().Begin()

	defer func() {
		if err == nil {
			db.Commit()
		} else {
			db.Rollback()
		}
	}()

	if info.Use == DockerSvrUse {
		err = db.Model(&DockerSvrInfo{}).Where("1=1").Update("use", DockerSvrNotUse).Error
		if err != nil {
			log.Println(err)
			panic(err)
		}
	}

	err = db.Create(info).Error
	if err != nil {
		log.Println(err)
		panic(err)
	}
}

func UpdateDockerSvrInfo(info *DockerSvrInfo) {
	var err error
	db := GetDb().Begin()

	defer func() {
		if err != nil {
			db.Rollback()
		} else {
			db.Commit()
		}
	}()

	if info.Use == DockerSvrUse {
		err = db.Model(&DockerSvrInfo{}).Where("1=1").Update("use", DockerSvrNotUse).Error
		if err != nil {
			log.Println(err)
			panic(err)
		}
	}

	err = db.Model(&DockerSvrInfo{}).Where("id = ?", info.Id).Updates(info).Error
	if err != nil {
		log.Println(err)
		panic(err)
	}
}

func DeleteDockerSvrInfo(id int) {
	err := GetDb().Delete(&DockerSvrInfo{}, id).Error
	if err != nil {
		log.Println(err)
		panic(err)
	}
}

func GetDockerSvrInfos() []DockerSvrInfo {
	var infos []DockerSvrInfo
	err := GetDb().Find(&infos).Error
	if err != nil {
		log.Println(err)
		panic(err)
	}

	return infos
}

func GetUseSvrId() int {
	info := GetUseSvrInfo()
	if info == nil {
		return 0
	}

	return info.Id
}

func GetUseSvrInfo() *DockerSvrInfo {
	var infos []DockerSvrInfo
	err := GetDb().Model(&DockerSvrInfo{}).Where("id!=0 and use=1").Limit(1).Find(&infos).Error
	if err != nil {
		log.Println(err)
		return nil
	}

	if len(infos) == 0 {
		return nil
	}

	return &infos[0]
}
