package docker

import (
	"context"
	"log"

	"github.com/docker/docker/api/types"
)

func GetDockerNasNetworkName() string {
	return "DockerNAS"
}

func IsNetworkExist() (bool, error) {
	ctx := context.Background()
	cli, err := ConnDocker()
	if err != nil {
		log.Println("create docker client error")
		return false, err
	}
	defer cli.Close()

	networks, err := cli.NetworkList(ctx, types.NetworkListOptions{})
	if err != nil {
		log.Println("create docker client error")
		return false, err
	}

	for _, network := range networks {
		if network.Name == GetDockerNasNetworkName() {
			return true, nil
		}
	}

	return false, nil
}

func CheckNetwork() error {
	isExist, err := IsNetworkExist()
	if err != nil {
		return err
	}
	if isExist {
		return nil
	}

	ctx := context.Background()
	cli, err := ConnDocker()
	if err != nil {
		log.Println("create docker client error")
		return err
	}
	defer cli.Close()

	_, err = cli.NetworkCreate(ctx, GetDockerNasNetworkName(), types.NetworkCreate{})
	return err
}
