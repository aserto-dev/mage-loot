package docker

import (
	"context"
	"fmt"
	"strings"

	"github.com/aserto-dev/clui"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
)

var (
	ui = clui.NewUI()
)

type Arg func(*dockerArgs)

type PublishedPort struct {
	ContainerPort string
	HostIP        string
	HostPort      string
}

type dockerArgs struct {
	envVars      []string
	capAdds      []string
	publishPorts []PublishedPort
}

func Run(image string, args ...Arg) error {
	dargs := &dockerArgs{}

	for _, arg := range args {
		arg(dargs)
	}

	ctx := context.Background()

	cfg := &config{
		containerConfig: &container.Config{
			Image:        image,
			Env:          dargs.envVars,
			ExposedPorts: publishedPortToPortSet(dargs.publishPorts),
		},
		containerHostConfig: &container.HostConfig{
			CapAdd:       dargs.capAdds,
			PortBindings: publishedPortToPortMap(dargs.publishPorts),
		},
	}

	containerName := containerNameFromImage(image)

	cli, err := newCLI(cfg, ui)
	if err != nil {
		return err
	}

	containerType, err := cli.getContainer(ctx, containerName)
	if err != nil {
		return err
	}

	if containerType != nil {
		err = cli.removeContainer(ctx, containerName)
		if err != nil {
			return err
		}
	}

	return cli.startContainer(ctx, containerName)

}

func publishedPortToPortMap(publishedPorts []PublishedPort) nat.PortMap {
	result := make(nat.PortMap)

	for _, publishedPort := range publishedPorts {
		natContainerPort := nat.Port(publishedPort.ContainerPort)
		hostIP := publishedPort.HostIP
		if hostIP == "" {
			hostIP = "127.0.0.1"
		}
		result[natContainerPort] = append(result[natContainerPort], nat.PortBinding{
			HostIP:   hostIP,
			HostPort: publishedPort.HostPort,
		})
	}

	return result
}

func publishedPortToPortSet(publishedPorts []PublishedPort) nat.PortSet {
	result := make(nat.PortSet)

	for _, publishedPort := range publishedPorts {
		result[nat.Port(publishedPort.ContainerPort)] = struct{}{}
	}
	return result
}

func containerNameFromImage(image string) string {
	image = strings.ReplaceAll(image, ":", "_")
	return fmt.Sprintf("mage-loot-%s", image)
}

func WithEnvVar(key, value string) func(*dockerArgs) {
	return func(o *dockerArgs) {
		o.envVars = append(o.envVars, fmt.Sprintf("%s=%s", key, value))
	}
}

func WithCappAdd(capStr string) func(*dockerArgs) {
	return func(o *dockerArgs) {
		o.capAdds = append(o.capAdds, capStr)
	}
}

func WithPublishedPort(publishedPort PublishedPort) func(*dockerArgs) {
	return func(o *dockerArgs) {
		o.publishPorts = append(o.publishPorts, publishedPort)
	}
}
