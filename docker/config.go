package docker

import (
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/registry"
)

type config struct {
	containerConfig     *container.Config
	containerHostConfig *container.HostConfig
	credentials         *registry.AuthConfig
	networkName         string
}
