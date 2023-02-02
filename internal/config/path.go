package config

import (
	"dockernas/internal/utils"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// func GetBasePath() string {
// 	basePath, err := os.Getwd()
// 	if err != nil {
// 		panic(err)
// 	}
// 	basePath = filepath.Join(basePath, "data")
// 	return basePath
// }

func GetBasePath() string {
	basePath := GetConfig("basePath", "")
	if basePath == "" {
		panic("base data path is not set")
	}
	return basePath
}

func IsBasePathSet() bool {
	return GetConfig("basePath", "") != ""
}

func SetBasePath(path string) {
	if IsBasePathSet() {
		panic("base data path has set")
	}
	SetConfig("basePath", path)
	SaveConfig()
	InitLogger()
}

func GetFullDfsPath(path string) string {
	basePath := GetBasePath()
	basePath = filepath.Join(basePath, "dfs", path)
	return basePath
}

func GetDBFilePath() string {
	basePath := GetBasePath()
	basePath = filepath.Join(basePath, "meta")
	utils.CheckCreateDir(basePath)
	return filepath.Join(basePath, "data.db3")
}

func GetExtraAppPath() string {
	basePath := GetBasePath()
	basePath = filepath.Join(basePath, "apps")
	utils.CheckCreateDir(basePath)
	return basePath
}

func GetAppLocalPath(instanceName string) string {
	basePath := GetBasePath()
	basePath = filepath.Join(basePath, "local", instanceName)
	return basePath
}

func GetAppLocalFilePath(instanceName string, fileName string) string {
	return filepath.Join(GetAppLocalPath(instanceName), fileName)
}

func GetLocalVolumePath(instanceName string, volumeName string) string {
	basePath := GetBasePath()
	basePath = filepath.Join(basePath, "local", instanceName, volumeName)
	utils.CheckCreateDir(basePath)
	return basePath
}

func GetAppMountFilePath(appName string, version string, fileName string) string {
	if strings.Index(appName, "/") > 0 {
		return filepath.Join(GetExtraAppPath(), appName, "docker", version, fileName)
	}

	dir1, err1 := ioutil.ReadDir("./apps")
	if err1 != nil {
		log.Println("list dir error", err1)
	} else {
		for _, fi1 := range dir1 {
			if fi1.IsDir() {
				dir2, err2 := ioutil.ReadDir(filepath.Join("./apps", fi1.Name()))
				if err2 != nil {
					log.Println("list dir error", err2)
				} else {
					for _, fi2 := range dir2 {
						if fi2.IsDir() {
							if fi2.Name() == appName {
								pwd, err := os.Getwd()
								if err != nil {
									panic(err)
								}
								return filepath.Join(pwd, "apps", fi1.Name(), appName, "docker", version, fileName)
							}
						}
					}
				}
			}
		}
	}

	panic("unkown app :" + appName)
}
