package docker

import (
	"DevTestSetup/internal/config"
	"context"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

func StartOpaContainer(ctx context.Context, cli *client.Client) (string, error) {
	err := RemoveExistingContainer(ctx, cli, config.CurrentDevTestSetupConfig.Opa.ContainerName)
	if err != nil {
		return "", nil
	}

	imageName := config.CurrentDevTestSetupConfig.Opa.Image + ":" + config.CurrentDevTestSetupConfig.Opa.Tag

	// Define host networking configuration.
	hostConfig := &container.HostConfig{
		PortBindings: nat.PortMap{
			"8181/tcp": []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: config.CurrentDevTestSetupConfig.Opa.Port,
				},
			},
		},
	}

	// Define container configuration
	cfg := &container.Config{
		Image: imageName,
		Cmd: []string{
			"run",
			"--server",
			"--log-level",
			"debug",
		},
		ExposedPorts: nat.PortSet{
			nat.Port(config.CurrentDevTestSetupConfig.Opa.Port + "/tcp"): struct{}{},
		},
	}

	return PullCreateStartContainer(ctx, cli, imageName, config.CurrentDevTestSetupConfig.Opa.ContainerName, hostConfig, cfg)
}

func StopOpaContainer(ctx context.Context, cli *client.Client) error {
	return StopContainerByName(ctx, cli, config.CurrentDevTestSetupConfig.Opa.ContainerName)
}
