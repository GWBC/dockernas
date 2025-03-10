package service

import (
	"dockernas/internal/backend/docker"
	"dockernas/internal/config"
	"dockernas/internal/models"
	"dockernas/internal/utils"
	"io/ioutil"
	"log"
	"strings"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
)

func GetHostInfo() models.HostInfo {
	var hostData models.HostInfo

	hostInfo, err := host.Info()
	if err != nil {
		panic(err)
	}
	cpuData, err := cpu.Info()
	if err != nil {
		panic(err)
	}
	memData, err := mem.VirtualMemory()
	if err != nil {
		panic(err)
	}

	hostData.HostName = hostInfo.Hostname
	hostData.DockerNASVersion = config.GetDockerNASVersion()
	hostData.BootTime = hostInfo.BootTime * 1000
	hostData.Platform = docker.DetectRealSystem()
	hostData.ModelName = cpuData[0].ModelName
	hostData.MemSize = memData.Total

	return hostData
}

func GetStorageInfo() models.StorageInfo {
	if !config.IsBasePathSet() {
		panic("base path is not set")
	}

	var storageInfo models.StorageInfo
	storageInfo.BaseDir = config.GetBasePath()

	if config.IsRunInConainer() {
		storageInfo.Device, storageInfo.Fstype, storageInfo.Capacity, storageInfo.FreeSize =
			utils.GetDeviceSizeByCmd(config.GetBasePath())
		storageInfo.BaseDir = docker.GetPathOnHost(config.GetBasePath())
	} else {
		infos, err := disk.Partitions(false)
		if err != nil {
			panic(err)
		}
		for _, info := range infos {
			log.Println(info)
			if strings.Index(storageInfo.BaseDir, info.Mountpoint) == 0 {
				storageInfo.Device = info.Device
				storageInfo.Fstype = info.Fstype
				diskUsage, error := disk.Usage(info.Mountpoint)
				if error != nil {
					panic(err)
				}
				storageInfo.Capacity = int64(diskUsage.Total)
				storageInfo.FreeSize = int64(diskUsage.Free)

				break
			}
		}
	}

	storageInfo.DfsSize, _ = utils.GetDirectorySize(config.GetFullDfsPath(""))
	storageInfo.InstanceSizeMap = map[string]int64{}
	dirs, _ := ioutil.ReadDir(config.GetAppLocalPath(""))
	for _, fi := range dirs {
		if fi.IsDir() {
			size, _ := utils.GetDirectorySize(config.GetAppLocalPath(fi.Name()))
			storageInfo.InstanceSizeMap[fi.Name()] = size
			storageInfo.LocalSize += size
		}
	}
	storageInfo.OtherSize = storageInfo.Capacity - storageInfo.FreeSize - storageInfo.LocalSize - storageInfo.DfsSize

	return storageInfo
}
