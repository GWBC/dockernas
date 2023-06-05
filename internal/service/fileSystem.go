package service

import (
	"dockernas/internal/backend/docker"
	"dockernas/internal/config"
	"dockernas/internal/models"
	"dockernas/internal/utils"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"

	"github.com/shirou/gopsutil/disk"
)

func getDirInfo(fullPath string, relativePath string) []models.DirInfo {
	dirInfoList := []models.DirInfo{}

	dirs, err := ioutil.ReadDir(fullPath)
	if err != nil {
		log.Println("list dir error", err)
		panic(err)
	}

	for _, fi := range dirs {
		if fi.IsDir() {
			var dirInfo models.DirInfo
			dirInfo.Name = fi.Name()
			dirInfo.Label = filepath.ToSlash(filepath.Join(relativePath, dirInfo.Name))
			dirInfo.Value = dirInfo.Label
			dirInfoList = append(dirInfoList, dirInfo)
		}
	}

	return dirInfoList
}

func getRemoteDir(fullPath string, relativePath string) []models.DirInfo {
	dirInfoList := []models.DirInfo{}

	rootPath := filepath.ToSlash(filepath.Join("/tmp", fullPath))

	strDirs := docker.HostMachineExec([]string{"find",
		rootPath,
		"-maxdepth", "1", "-type", "d"})
	dirs := strings.Split(string(strDirs), "\n")

	for _, dir := range dirs {
		if len(dir) == 0 || dir == rootPath {
			continue
		}

		var dirInfo models.DirInfo
		dirInfo.Name = filepath.Base(dir)
		dirInfo.Label = filepath.ToSlash(filepath.Join(relativePath, dirInfo.Name))
		dirInfo.Value = dirInfo.Label
		dirInfoList = append(dirInfoList, dirInfo)
	}

	return dirInfoList
}

func GetDfsDirInfo(path string) []models.DirInfo {
	basePath := config.GetFullDfsPath(path)

	if models.GetUseSvrId() == 0 {
		utils.CheckCreateDir(config.GetFullDfsPath(""))
		return getDirInfo(basePath, path)
	}

	return getRemoteDir(basePath, path)
}

func GetSystemDirInfo(path string) []models.DirInfo {
	if path == "" {
		dirInfoList := []models.DirInfo{}
		if utils.IsRunOnWindows() {
			infos, err := disk.Partitions(false)
			if err != nil {
				panic(err)
			}
			for _, info := range infos {
				var dirInfo models.DirInfo
				dirInfo.Name = info.Mountpoint
				if !strings.HasSuffix(info.Mountpoint, "/") {
					dirInfo.Label = info.Mountpoint + "/"
					dirInfo.Value = info.Mountpoint + "/"
				} else {
					dirInfo.Label = info.Mountpoint
					dirInfo.Value = info.Mountpoint
				}
				dirInfoList = append(dirInfoList, dirInfo)
			}
		} else {
			var dirInfo models.DirInfo
			dirInfo.Name = "/"
			dirInfo.Label = "/"
			dirInfo.Value = "/"
			dirInfoList = append(dirInfoList, dirInfo)
		}
		return dirInfoList
	} else {
		return getDirInfo(path, path)
	}

}

func SetBasePath(path string) {
	config.SetBasePath(path)
}
