package docker

import (
	"DevTestSetup/internal/config"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
	"io"
)

func PullCreateStartContainer(ctx context.Context, cli *client.Client, imageName string, containerName string, hostConfig *container.HostConfig, cfg *container.Config) (string, error) {
	// Prepare the networking configuration
	networkConfig := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			"intranet": {},
		},
	}

	if config.CurrentDevTestSetupConfig.Docker.RepoUrl != "" {
		imageName = config.CurrentDevTestSetupConfig.Docker.RepoUrl + "/" + imageName
	}

	opts, err := authOptions()
	if err != nil {
		return "", err
	}

	// pulling the image
	pullResp, err := cli.ImagePull(ctx, imageName, opts)
	if err != nil {
		err = fmt.Errorf("error occurred while pulling the image %s: %w", imageName, err)
		return "", err
	}
	defer pullResp.Close()

	// Just read from the Pull response to avoid issues with 'image not pulled'
	if _, err := io.Copy(io.Discard, pullResp); err != nil {
		err = fmt.Errorf("error occurred while pulling the image %s: %w", imageName, err)
	}

	resp, err := cli.ContainerCreate(ctx, cfg, hostConfig, networkConfig, nil, containerName)
	if err != nil {
		err = fmt.Errorf("error occurred while creating container %s: %w", containerName, err)
		return "", err
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		err = fmt.Errorf("error occurred while starting container %s: %w", containerName, err)
		return "", err
	}

	return resp.ID, nil
}

func RemoveExistingContainer(ctx context.Context, cli *client.Client, containerName string) error {
	nameFilter := filters.NewArgs()
	nameFilter.Add("name", containerName)

	// Check if container exists
	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{All: true, Filters: nameFilter})
	if err != nil {
		err = fmt.Errorf("error occurred while getting list of containers: %w", err)
		return err
	}

	existingContainerID := ""
	for _, cont := range containers {
		for _, name := range cont.Names {
			if name == "/"+containerName {
				existingContainerID = cont.ID
				break
			}
		}
	}

	if existingContainerID != "" {
		log.Info("Container already exists, removing...")
		err = cli.ContainerRemove(ctx, existingContainerID, types.ContainerRemoveOptions{})
		if err != nil {
			err = fmt.Errorf("error occurred while removing existing container: %w", err)
			return err
		}
	}

	return nil
}

func StopContainerByName(ctx context.Context, cli *client.Client, containerName string) error {
	nameFilter := filters.NewArgs()
	// Filter docker containers based on the name
	nameFilter.Add("name", containerName)

	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{All: true, Filters: nameFilter})
	if err != nil {
		return err
	}
	for _, cont := range containers {
		err := cli.ContainerStop(ctx, cont.ID, container.StopOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}

func CreateNetworkAndVolumes(ctx context.Context, cli *client.Client) error {
	// Create Network
	networkCreateResponse, err := cli.NetworkCreate(ctx, "intranet", types.NetworkCreate{})
	if err != nil {
		err = fmt.Errorf("failed to create network: %w", err)
		return err
	}
	log.Info("Created Network: " + networkCreateResponse.ID)

	// Create Volume
	volumeCreateBody := volume.CreateOptions{
		Driver: "local",
		Name:   "redis",
	}
	volumeCreateResponse, err := cli.VolumeCreate(ctx, volumeCreateBody)
	if err != nil {
		err = fmt.Errorf("failed to create volume: %w", err)
		return err
	}
	log.Info("Created volume: " + volumeCreateResponse.Name)

	return nil
}

func RemoveNetwork(ctx context.Context, cli *client.Client) error {
	err := cli.NetworkRemove(ctx, "intranet")
	if err != nil {
		err = fmt.Errorf("failed to remove network: %w", err)
		return err
	}
	log.Info("Removed network: intranet")

	return nil
}

func authOptions() (opts types.ImagePullOptions, err error) {
	opts = types.ImagePullOptions{}
	err = nil

	if config.CurrentDevTestSetupConfig.Docker.Username != "" && config.CurrentDevTestSetupConfig.Docker.Password != "" {
		authConfig := registry.AuthConfig{
			Username: config.CurrentDevTestSetupConfig.Docker.Username,
			Password: config.CurrentDevTestSetupConfig.Docker.Password,
		}
		encodedJSON, err := json.Marshal(authConfig)
		if err != nil {
			err = fmt.Errorf("error occurred while setting up authentication: %w", err)
		}
		opts.RegistryAuth = base64.URLEncoding.EncodeToString(encodedJSON)
	}

	return opts, err
}
