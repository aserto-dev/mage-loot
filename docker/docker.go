package docker

import (
	"context"
	"fmt"
	"strings"

	"github.com/aserto-dev/clui"
	"github.com/docker/docker/api/types/container"
)

var (
	ui = clui.NewUI()
)

type Arg func(*dockerArgs)

type dockerArgs struct {
	envVars []string
	capAdds []string
}

func Run(image string, args ...Arg) error {
	dargs := &dockerArgs{}

	for _, arg := range args {
		arg(dargs)
	}

	ctx := context.Background()

	cfg := &config{
		containerConfig: &container.Config{
			Image: image,
			Env:   dargs.envVars,
		},
		containerHostConfig: &container.HostConfig{
			CapAdd: dargs.capAdds,
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

func containerNameFromImage(image string) string {
	image = strings.ReplaceAll(image, ":", "_")
	return fmt.Sprintf("mage-loot-%s", image)
}

func WithEnvVar(key, value string) func(*dockerArgs) {
	return func(o *dockerArgs) {
		o.envVars = append(o.envVars, fmt.Sprintf("%s=%s", key, value))
	}
}

func WithCappAdd(cap string) func(*dockerArgs) {
	return func(o *dockerArgs) {
		o.envVars = append(o.capAdds, cap)
	}
}
