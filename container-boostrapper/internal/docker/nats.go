package docker

import (
	"DevTestSetup/internal/config"
	"context"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

func StartNatsContainer(ctx context.Context, cli *client.Client) (string, error) {
	err := RemoveExistingContainer(ctx, cli, config.CurrentDevTestSetupConfig.Nats.ContainerName)
	if err != nil {
		return "", nil
	}

	imageName := config.CurrentDevTestSetupConfig.Nats.Image + ":" + config.CurrentDevTestSetupConfig.Nats.Tag

	// Define host networking configuration.
	hostConfig := &container.HostConfig{
		PortBindings: nat.PortMap{
			"4222/tcp": []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: config.CurrentDevTestSetupConfig.Nats.ServerPort,
				},
			},
			"8222/tcp": []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: config.CurrentDevTestSetupConfig.Nats.MonitoringPort,
				},
			},
		},
	}

	// Define container configuration
	cfg := &container.Config{
		Image: imageName,
		Cmd: []string{
			"--cluster_name",
			config.CurrentDevTestSetupConfig.Nats.ClusterName,
			"--cluster",
			"nats://0.0.0.0:6222",
			"--http_port",
			config.CurrentDevTestSetupConfig.Nats.MonitoringPort,
		},
		ExposedPorts: nat.PortSet{
			nat.Port(config.CurrentDevTestSetupConfig.Nats.ServerPort + "/tcp"):     struct{}{},
			nat.Port(config.CurrentDevTestSetupConfig.Nats.MonitoringPort + "/tcp"): struct{}{},
		},
	}

	return PullCreateStartContainer(ctx, cli, imageName, config.CurrentDevTestSetupConfig.Nats.ContainerName, hostConfig, cfg)
}

func StopNatsContainer(ctx context.Context, cli *client.Client) error {
	return StopContainerByName(ctx, cli, config.CurrentDevTestSetupConfig.Nats.ContainerName)
}
