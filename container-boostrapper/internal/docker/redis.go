package docker

import (
	"DevTestSetup/internal/config"
	"context"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

func StartRedisContainer(ctx context.Context, cli *client.Client) (string, error) {
	err := RemoveExistingContainer(ctx, cli, config.CurrentDevTestSetupConfig.Redis.ContainerName)
	if err != nil {
		return "", nil
	}

	imageName := config.CurrentDevTestSetupConfig.Redis.Image + ":" + config.CurrentDevTestSetupConfig.Redis.Tag

	// Define host networking configuration.
	hostConfig := &container.HostConfig{
		PortBindings: nat.PortMap{
			"6379/tcp": []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: config.CurrentDevTestSetupConfig.Redis.Port,
				},
			},
		},
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeVolume,
				Source: "redis",
				Target: "/data",
			},
		},
		NetworkMode: "intranet",
	}

	// Define container configuration
	cfg := &container.Config{
		Image: imageName,
		Cmd: []string{
			"redis-server",
			"--save",
			"20",
			"1",
			"--loglevel",
			"warning",
			"--requirepass",
			"r7fOGyA5Ve",
		},
		ExposedPorts: nat.PortSet{
			nat.Port(config.CurrentDevTestSetupConfig.Redis.Port + "/tcp"): struct{}{},
		},
	}

	return PullCreateStartContainer(ctx, cli, imageName, config.CurrentDevTestSetupConfig.Redis.ContainerName, hostConfig, cfg)
}

func StopRedisContainer(ctx context.Context, cli *client.Client) error {
	return StopContainerByName(ctx, cli, config.CurrentDevTestSetupConfig.Redis.ContainerName)
}
