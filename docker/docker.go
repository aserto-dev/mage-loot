package docker

import (
	"context"
	"fmt"
	"strings"

	"github.com/aserto-dev/clui"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/registry"
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
	envVars       []string
	capAdds       []string
	publishPorts  []PublishedPort
	registryCreds *registry.AuthConfig
	entrypoint    []string
	networkName   string
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
			Entrypoint:   dargs.entrypoint,
		},
		containerHostConfig: &container.HostConfig{
			CapAdd:       dargs.capAdds,
			PortBindings: publishedPortToPortMap(dargs.publishPorts),
		},
		credentials: dargs.registryCreds,
		networkName: dargs.networkName,
	}

	containerName := sanitizeName(image)

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

func GetContainer(containerName string) (*types.Container, error) {
	ctx := context.Background()

	cfg := &config{}

	containerName = sanitizeName(containerName)

	cli, err := newCLI(cfg, ui)
	if err != nil {
		return nil, err
	}

	return cli.getContainer(ctx, containerName)
}

func CreateNetwork(name string) (string, error) {
	ctx := context.Background()

	cfg := &config{}

	cli, err := newCLI(cfg, ui)
	if err != nil {
		return "", err
	}

	nwName := sanitizeName(name)
	return cli.createNetwork(ctx, nwName)
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

func sanitizeName(name string) string {
	name = strings.ReplaceAll(name, ":", "_")
	name = strings.ReplaceAll(name, "/", "_")
	return fmt.Sprintf("mage-loot-%s", name)
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

func WithCredentials(credentials *registry.AuthConfig) func(*dockerArgs) {
	return func(o *dockerArgs) {
		o.registryCreds = credentials
	}
}

func WithEntrypoint(entrypoint []string) func(*dockerArgs) {
	return func(o *dockerArgs) {
		o.entrypoint = entrypoint
	}
}

func WithNetwork(networkName string) func(*dockerArgs) {
	return func(o *dockerArgs) {
		o.networkName = networkName
	}
}
