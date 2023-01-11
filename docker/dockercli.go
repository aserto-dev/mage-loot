package docker

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"

	"github.com/aserto-dev/clui"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type dockerCLI struct {
	dockerClient *client.Client
	ui           *clui.UI
	cfg          *config
}

func newCLI(cfg *config, ui *clui.UI) (*dockerCLI, error) {
	ui.Note().Msg("initializing docker client")
	cli, err := client.NewClientWithOpts()
	if err != nil {
		return nil, errors.Wrap(err, "failed to start docker client")
	}

	return &dockerCLI{
		dockerClient: cli,
		ui:           ui,
		cfg:          cfg,
	}, nil
}

func (cli *dockerCLI) startContainer(ctx context.Context, containerName string) error {
	image := cli.cfg.containerConfig.Image
	cli.ui.Note().Msgf("checking image %s", image)
	imagePullOptions := types.ImagePullOptions{}
	if cli.cfg.credentials != nil {
		encodedJSON, err := json.Marshal(cli.cfg.credentials)
		if err != nil {
			return err
		}
		authStr := base64.URLEncoding.EncodeToString(encodedJSON)
		imagePullOptions.RegistryAuth = authStr
	}

	ioReader, err := cli.dockerClient.ImagePull(ctx, image, imagePullOptions)
	if err != nil {
		return errors.Wrapf(err, "failed to pull image [%s]", image)
	}
	_, err = io.Copy(cli.ui.Output(), ioReader)
	if err != nil {
		return err
	}

	defer ioReader.Close()

	ui.Note().Msg("creating docker container")

	var networkConfig *network.NetworkingConfig
	if cli.cfg.networkName != "" {
		endpoints := make(map[string]*network.EndpointSettings, 1)
		endpoints[cli.cfg.networkName] = &network.EndpointSettings{}
		networkConfig = &network.NetworkingConfig{
			EndpointsConfig: endpoints,
		}
	}

	resp, err := cli.dockerClient.ContainerCreate(
		ctx,
		cli.cfg.containerConfig,
		cli.cfg.containerHostConfig,
		networkConfig,
		nil,
		containerName,
	)
	if err != nil {
		return errors.Wrap(err, "failed to create container")
	}

	startupOptions := types.ContainerStartOptions{}

	ui.Note().Msg("starting docker container")
	err = cli.dockerClient.ContainerStart(ctx, resp.ID, startupOptions)
	if err != nil {
		return errors.Wrap(err, "failed to start container")
	}

	ui.Note().Msg("checking if the container is running")
	containerType, err := cli.getContainer(ctx, containerName)
	if err != nil {
		return errors.Wrap(err, "failed to get container")
	}

	if containerType.State != "running" {
		cli.ui.Problem().Msgf("container [%s] is not running, check logs by running 'docker logs %s'", containerName, containerType.ID)
		return fmt.Errorf("failed to start container")
	}

	return nil
}

func (cli *dockerCLI) removeContainer(ctx context.Context, containerName string) error {
	cli.ui.Note().Msgf("removing container with name [%s]", containerName)
	err := cli.dockerClient.ContainerRemove(ctx, containerName, types.ContainerRemoveOptions{
		Force:         true,
		RemoveVolumes: true,
	})

	return errors.Wrap(err, "failed to remove container")
}

func (cli *dockerCLI) getContainer(ctx context.Context, containerName string) (*types.Container, error) {
	containerTypes, err := cli.dockerClient.ContainerList(ctx, types.ContainerListOptions{
		All:     true,
		Filters: filters.NewArgs(filters.Arg("name", containerName)),
	})

	if err != nil {
		return nil, err
	}

	if len(containerTypes) == 0 {
		return nil, nil
	}
	return &containerTypes[0], nil
}

func (cli *dockerCLI) createNetwork(ctx context.Context, name string) (string, error) {
	networks, err := cli.dockerClient.NetworkList(ctx,
		types.NetworkListOptions{Filters: filters.NewArgs(filters.KeyValuePair{Key: "name", Value: name})})
	if err != nil {
		return "", errors.Wrap(err, "failed to read networks")
	}

	for i := range networks {
		n := networks[i]
		err := cli.dockerClient.NetworkRemove(ctx, n.ID)
		if err != nil {
			log.Error().Err(err)
		}
	}

	net, err := cli.dockerClient.NetworkCreate(ctx, name, types.NetworkCreate{
		CheckDuplicate: false,
		Driver:         "bridge",
	})
	return net.ID, err
}
