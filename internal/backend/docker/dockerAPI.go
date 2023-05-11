package docker

import (
	"bytes"
	"context"
	"dockernas/internal/config"
	"dockernas/internal/models"
	"dockernas/internal/utils"
	"io"
	"log"
	"math"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

func ConnDocker() (*client.Client, error) {
	info := models.GetUseSvrInfo()
	if info == nil || len(info.IP) == 0 {
		return client.NewClientWithOpts(client.FromEnv,
			client.WithAPIVersionNegotiation())
	}

	return client.NewClientWithOpts(client.FromEnv,
		client.WithAPIVersionNegotiation(),
		client.WithHost("tcp://"+info.IP))
}

func CreateAssistContainer(containerName string) (string, error) {
	exist, containerId := IsContainerNameExist(containerName)
	if exist {
		return containerId, nil
	}

	ctx := context.Background()
	cli, err := ConnDocker()
	if err != nil {
		log.Println("create docker client error")
		return containerId, err
	}
	defer cli.Close()

	imageUrl := "busybox"

	if !IsImageExist(imageUrl) {
		reader, err := cli.ImagePull(ctx, imageUrl, types.ImagePullOptions{})
		if err != nil {
			log.Println("pull image error: " + imageUrl)
			return containerId, err
		}

		defer reader.Close()
	}

	containerConfig := container.Config{
		Image:        imageUrl,
		ExposedPorts: make(nat.PortSet),
		Env:          []string{},
		Cmd:          []string{"sh", "-c", "while [ 1 ]; do sleep 3600; done"},
	}

	hostConfig := container.HostConfig{
		PortBindings: make(nat.PortMap),
		Mounts: []mount.Mount{mount.Mount{
			Type:   mount.TypeBind,
			Source: "/",
			Target: "/tmp",
		}},
		RestartPolicy: container.RestartPolicy{Name: "always"},
		NetworkMode:   "none",
		Privileged:    false,
	}

	resp, err := cli.ContainerCreate(ctx, &containerConfig, &hostConfig, nil, nil, containerName)
	if err != nil {
		log.Println("create container error")
		return containerId, err
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		log.Println("run container error")
		return resp.ID, err
	}

	return resp.ID, nil
}

func HostMachineMakeDir(path string) error {
	containerName := "dockernas-assist"
	_, err := CreateAssistContainer(containerName)
	if err != nil {
		return err
	}

	ctx := context.Background()
	cli, err := ConnDocker()
	if err != nil {
		log.Println("create docker client error")
		return err
	}
	defer cli.Close()

	dirPath := filepath.ToSlash(filepath.Join("/tmp", path))

	//创建执行环境
	ir, err := cli.ContainerExecCreate(ctx, containerName, types.ExecConfig{
		Cmd: []string{"mkdir", "-p", dirPath},
	})

	//执行
	err = cli.ContainerExecStart(ctx, ir.ID, types.ExecStartCheck{Detach: false, Tty: true})
	if err != nil {
		log.Println("exec command error")
		panic(err)
	}

	//创建执行环境
	ir, err = cli.ContainerExecCreate(ctx, containerName, types.ExecConfig{
		Cmd: []string{"chmod", "777", dirPath},
	})

	//执行
	err = cli.ContainerExecStart(ctx, ir.ID, types.ExecStartCheck{Detach: false, Tty: true})
	if err != nil {
		log.Println("exec command error")
		panic(err)
	}

	return err
}

func Delete(containerID string) error {
	ctx := context.Background()
	cli, err := ConnDocker()
	if err != nil {
		log.Println("create docker client error")
		return err
	}
	defer cli.Close()

	err = cli.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{Force: true})
	if err != nil {
		log.Println("start docker error")
		return err
	}

	return nil
}

func Stop(containerID string) error {
	ctx := context.Background()
	cli, err := ConnDocker()
	if err != nil {
		log.Println("create docker client error")
		return err
	}
	defer cli.Close()

	timeoutSecond := time.Second * 10
	err = cli.ContainerStop(ctx, containerID, &timeoutSecond)

	if err != nil {
		log.Println("stop docker error")
		return err
	}

	return nil
}

func Start(containerID string) error {
	ctx := context.Background()
	cli, err := ConnDocker()
	if err != nil {
		log.Println("create docker client error")
		return err
	}
	defer cli.Close()

	err = cli.ContainerStart(ctx, containerID, types.ContainerStartOptions{})
	if err != nil {
		log.Println("start docker error")
		return err
	}

	return nil
}

func Restart(containerID string) error {
	ctx := context.Background()
	cli, err := ConnDocker()
	if err != nil {
		log.Println("create docker client error")
		return err
	}
	defer cli.Close()

	timeoutSecond := time.Second * 10
	err = cli.ContainerRestart(ctx, containerID, &timeoutSecond)
	if err != nil {
		log.Println("restart docker error")
		return err
	}

	return nil
}

func PullImage(imageUrl string) (io.ReadCloser, error) {
	ctx := context.Background()
	cli, err := ConnDocker()
	if err != nil {
		log.Println("create docker client error")
		return nil, err
	}
	defer cli.Close()

	reader, err2 := cli.ImagePull(ctx, imageUrl, types.ImagePullOptions{})
	if err2 != nil {
		log.Println("pull image error: " + imageUrl)
		return nil, err2
	}

	return reader, nil
}

func DelImage(imageId string) {
	ctx := context.Background()
	cli, err := ConnDocker()
	if err != nil {
		log.Println("create docker client error")
		panic(err)
	}
	defer cli.Close()

	_, err2 := cli.ImageRemove(ctx, imageId, types.ImageRemoveOptions{})
	if err2 != nil {
		panic(err2)
	}
}

func ListImage() []models.ImageInfo {
	ctx := context.Background()
	cli, err := ConnDocker()
	if err != nil {
		log.Println("create docker client error")
		panic(err)
	}
	defer cli.Close()

	images, err2 := cli.ImageList(ctx, types.ImageListOptions{All: true})
	if err2 != nil {
		panic(err2)
	}

	var infos []models.ImageInfo
	for _, v := range images {
		for _, tag := range v.RepoTags {
			infos = append(infos, models.ImageInfo{
				Id:    v.ID,
				Name:  tag,
				Size:  v.Size,
				State: "100%",
			})
		}
	}

	return infos
}

func IsImageExist(name string) bool {
	for _, img := range ListImage() {
		if strings.Contains(img.Name, name) {
			return true
		}
	}
	return false
}

func ListContainer() []types.Container {
	ctx := context.Background()
	cli, err := ConnDocker()
	if err != nil {
		log.Println("create docker client error")
		panic(err)
	}
	defer cli.Close()

	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{All: true})
	if err != nil {
		panic(err)
	}

	return containers
}

func IsContainerExist(containerID string) bool {
	for _, container := range ListContainer() {
		if container.ID == containerID {
			return true
		}
	}
	return false
}

func IsContainerNameExist(name string) (bool, string) {
	for _, container := range ListContainer() {
		for _, cname := range container.Names {
			if cname == "/"+name {
				return true, container.ID
			}
		}
	}
	return false, ""
}

func GetContainerInspect(containerID string) types.ContainerJSON {
	ctx := context.Background()
	cli, err := ConnDocker()
	if err != nil {
		log.Println("create docker client error")
		panic(err)
	}
	defer cli.Close()

	data, err := cli.ContainerInspect(ctx, containerID)
	if err != nil {
		panic(err)
	}

	return data
}

func Create(param *models.InstanceParam) (string, error) {
	containerConfig, hostConfig := buildConfig(param)

	err := CheckNetwork()
	if err != nil {
		log.Println("check network error")
		return "", err
	}

	ctx := context.Background()
	cli, err := ConnDocker()
	if err != nil {
		log.Println("create docker client error")
		return "", err
	}
	defer cli.Close()

	// _, err = cli.ImagePull(ctx, param.ImageUrl, types.ImagePullOptions{})
	// if err != nil {
	// 	log.Println("pull image error: " + param.ImageUrl)
	// 	return "", err
	// }

	resp, err := cli.ContainerCreate(ctx, &containerConfig, &hostConfig, nil, nil, param.Name)
	if err != nil {
		log.Println("create container error")
		return "", err
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		log.Println("run container error")
		return resp.ID, err
	}

	return resp.ID, nil
}

func GetContainerStat(id string) (types.ContainerStats, error) {
	ctx := context.Background()
	cli, err := ConnDocker()
	if err != nil {
		log.Println("create docker client error")
		panic(err)
	}
	defer cli.Close()

	return cli.ContainerStats(ctx, id, false)
}

func replaceVariable(aStr string, param *models.InstanceParam) string {
	if param.OtherParams == nil {
		return aStr
	}

	for _, v := range param.OtherParams {
		if v.OtherType == models.PLACEHOLDER_PARAM {
			aStr = strings.ReplaceAll(aStr, v.Key, v.Value)
		}
	}

	return aStr
}

func buildConfig(param *models.InstanceParam) (container.Config, container.HostConfig) {
	m := make([]mount.Mount, 0, len(param.DfsVolume)+len(param.LocalVolume))

	useSvrId := models.GetUseSvrId()

	var usedVolumeName []string
	for index, item := range param.LocalVolume {
		if item.Name == "" || utils.Contains(usedVolumeName, item.Name) {
			panic("local volume name error:" + item.Name)
		}
		usedVolumeName = append(usedVolumeName, item.Name)

		if item.MountFile {
			localDir := config.GetLocalVolumePath(param.Name, "")
			if useSvrId == 0 {
				utils.CheckCreateDir(localDir)
			} else {
				HostMachineMakeDir(localDir)
			}

			templateFilePath := config.GetAppMountFilePath(param.Path, item.Name)
			instanceLocalPath := config.GetAppLocalFilePath(param.Name, item.Name)
			if !utils.IsFileExist(instanceLocalPath) {
				_, err := utils.CopyFile(templateFilePath, instanceLocalPath)
				utils.MakePathReadAble(instanceLocalPath)
				if err != nil {
					panic(err)
				}
			}
			param.LocalVolume[index].Value = instanceLocalPath
			m = append(m, mount.Mount{
				Type:   mount.TypeBind,
				Source: GetPathOnHost(instanceLocalPath),
				Target: item.Key,
			})
		} else {
			localDir := config.GetLocalVolumePath(param.Name, item.Name)
			if useSvrId == 0 {
				utils.CheckCreateDir(localDir)
				utils.MakePathReadAble(localDir)
			} else {
				HostMachineMakeDir(localDir)
			}

			param.LocalVolume[index].Value = localDir
			m = append(m, mount.Mount{
				Type:   mount.TypeBind,
				Source: GetPathOnHost(localDir),
				Target: item.Key,
			})
		}
	}

	for index, item := range param.DfsVolume {
		//mount a local dir if dfs dir is empty, let user decide whether or not delete data when delete instance
		if item.Value == "" {
			if item.Name == "" || utils.Contains(usedVolumeName, item.Name) {
				panic("local volume name error:" + item.Name)
			}
			usedVolumeName = append(usedVolumeName, item.Name)
			localDir := config.GetLocalVolumePath(param.Name, item.Name)
			if useSvrId == 0 {
				utils.CheckCreateDir(localDir)
				utils.MakePathReadAble(localDir)
			} else {
				HostMachineMakeDir(localDir)
			}

			m = append(m, mount.Mount{
				Type:   mount.TypeBind,
				Source: GetPathOnHost(localDir),
				Target: item.Key,
			})
		} else {
			if item.Value[0] != '/' {
				item.Value = "/" + item.Value
				param.DfsVolume[index].Value = item.Value
			}

			dfsPath := config.GetFullDfsPath(item.Value)

			if useSvrId == 0 {
				utils.CheckCreateDir(dfsPath)
				utils.MakePathReadAble(dfsPath)
			} else {
				HostMachineMakeDir(dfsPath)
			}

			m = append(m, mount.Mount{
				Type:   mount.TypeBind,
				Source: GetPathOnHost(dfsPath),
				Target: item.Key,
			})
		}
	}

	var envs []string
	for _, item := range param.EnvParams {
		envs = append(envs, replaceVariable(item.Key, param)+"="+replaceVariable(item.Value, param))
	}

	cmdStr := replaceVariable(param.Cmd, param)
	var cmds []string
	if cmdStr != "" {
		cmds = strings.Split(cmdStr, " ")
	}

	exports := make(nat.PortSet)
	netPort := make(nat.PortMap)

	if param.NetworkMode != models.HOST_MODE && param.NetworkMode != models.NOBUND_MODE {
		hostIp := "0.0.0.0"
		if param.NetworkMode == models.LOCAL_MODE {
			hostIp = "127.0.0.1"
		}
		for _, item := range param.PortParams {
			if item.Value == "" {
				continue
			}
			proto := "tcp"
			if item.Protocol == "udp" {
				proto = "udp"
			}

			sportList := make([]int, 0)
			sports := strings.Split(item.Key, "-")
			if len(sports) == 2 {
				start, _ := strconv.Atoi(sports[0])
				end, _ := strconv.Atoi(sports[1])
				for i := start; i <= end; i++ {
					sportList = append(sportList, i)
				}
			} else {
				start, _ := strconv.Atoi(sports[0])
				sportList = append(sportList, start)
			}

			dportList := make([]int, 0)
			dports := strings.Split(item.Value, "-")
			if len(sports) == 2 {
				start, _ := strconv.Atoi(dports[0])
				end, _ := strconv.Atoi(dports[1])
				for i := start; i <= end; i++ {
					dportList = append(dportList, i)
				}
			} else {
				start, _ := strconv.Atoi(dports[0])
				dportList = append(dportList, start)
			}

			mlen := int(math.Min(float64(len(sportList)), float64(len(dportList))))
			for i := 0; i < mlen; i++ {
				natPort, _ := nat.NewPort(proto, strconv.Itoa(sportList[i]))
				exports[natPort] = struct{}{}

				portList := make([]nat.PortBinding, 0, 1)
				portList = append(portList, nat.PortBinding{HostIP: hostIp, HostPort: strconv.Itoa(dportList[i])})
				netPort[natPort] = portList
			}
		}
	}

	containerConfig := container.Config{
		Image:        param.ImageUrl,
		ExposedPorts: exports,
		Env:          envs,
		Cmd:          cmds,
	}
	hostConfig := container.HostConfig{
		PortBindings:  netPort,
		Mounts:        m,
		RestartPolicy: container.RestartPolicy{Name: "always"},
	}

	if param.NetworkMode != models.HOST_MODE {
		hostConfig.NetworkMode = container.NetworkMode(GetDockerNasNetworkName())
	} else {
		hostConfig.NetworkMode = "host"
	}

	if param.Privileged {
		hostConfig.Privileged = true
	}

	if param.User != "" {
		containerConfig.User = param.User
	}

	if DetectRealSystem() == "linux" {
		// if param.Privileged == false {
		// 	curUser, err := user.Current()
		// 	if err != nil {
		// 		panic("get current user error: " + err.Error())
		// 	}
		// 	containerConfig.User = curUser.Uid
		// }
		hostConfig.ExtraHosts = append(hostConfig.ExtraHosts, "host.docker.internal:host-gateway")
	}

	return containerConfig, hostConfig
}

func GetLog(containerID string) string {
	ctx := context.Background()
	cli, err := ConnDocker()
	if err != nil {
		log.Println("create docker client error")
		panic(err)
	}
	defer cli.Close()

	out, err := cli.ContainerLogs(ctx, containerID, types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		log.Println("get docker log error")
		panic(err)
	}
	defer out.Close()

	var writer bytes.Buffer
	io.Copy(&writer, out)

	return writer.String()
}

func GetDockerVersion() types.Version {
	ctx := context.Background()
	cli, err := ConnDocker()
	if err != nil {
		log.Println("create docker client error")
		panic(err)
	}
	defer cli.Close()

	version, err := cli.ServerVersion(ctx)
	if err != nil {
		log.Println("get docker version error")
		panic(err)
	}

	return version
}

func DetectRealSystem() string {
	version := GetDockerVersion()
	if strings.Contains(version.KernelVersion, "microsoft") &&
		strings.Contains(version.KernelVersion, "WSL") {
		return "windows"
	}
	return utils.GetOperationSystemName()
}

func Exec(container string, rows string, columns string) types.HijackedResponse {
	ctx := context.Background()
	cli, err := ConnDocker()
	if err != nil {
		log.Println("create docker client error")
		panic(err)
	}
	defer cli.Close()

	cmds := []string{"bash", "sh"}

	err = nil
	var ir types.IDResponse

	for _, cmd := range cmds {
		ir, err = cli.ContainerExecCreate(ctx, container, types.ExecConfig{
			AttachStdin:  true,
			AttachStdout: true,
			AttachStderr: true,
			Cmd:          []string{cmd},
			Env:          []string{"LINES=" + rows, "COLUMNS=" + columns, "TERM=xterm"},
			Tty:          true,
		})

		if err == nil {
			break
		}
	}

	if err != nil {
		log.Println("exec cmd error")
		panic(err)
	}

	// 附加到上面创建的/bin/bash进程中
	hr, err := cli.ContainerExecAttach(ctx, ir.ID, types.ExecStartCheck{Detach: false, Tty: true})
	if err != nil {
		log.Println("attch container error")
		panic(err)
	}

	return hr
}
