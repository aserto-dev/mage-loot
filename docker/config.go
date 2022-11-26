package docker

import "github.com/docker/docker/api/types/container"

type config struct {
	containerConfig     *container.Config
	containerHostConfig *container.HostConfig
}
