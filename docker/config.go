package docker

import (
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
)

type config struct {
	containerConfig     *container.Config
	containerHostConfig *container.HostConfig
	credentials         *types.AuthConfig
	networkName         string
}
