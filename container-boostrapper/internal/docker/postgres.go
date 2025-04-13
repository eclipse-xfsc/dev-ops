package docker

import (
	"DevTestSetup/internal/config"
	"context"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

func StartPostgresContainer(ctx context.Context, cli *client.Client) (string, error) {
	err := RemoveExistingContainer(ctx, cli, config.CurrentDevTestSetupConfig.Postgres.ContainerName)
	if err != nil {
		return "", nil
	}

	imageName := config.CurrentDevTestSetupConfig.Postgres.Image + ":" + config.CurrentDevTestSetupConfig.Postgres.Tag

	// Define host networking configuration.
	hostConfig := &container.HostConfig{
		PortBindings: nat.PortMap{
			"5432/tcp": []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: config.CurrentDevTestSetupConfig.Postgres.Port,
				},
			},
		},
		NetworkMode: "intranet",
	}

	// Define container configuration
	cfg := &container.Config{
		Image: imageName,
		Env: []string{
			"POSTGRES_USER=" + config.CurrentDevTestSetupConfig.Postgres.User,
			"POSTGRES_PASSWORD=" + config.CurrentDevTestSetupConfig.Postgres.Password,
			"POSTGRES_DB=" + config.CurrentDevTestSetupConfig.Postgres.Db,
		},
		ExposedPorts: nat.PortSet{
			nat.Port(config.CurrentDevTestSetupConfig.Postgres.Port + "/tcp"): struct{}{},
		},
	}

	return PullCreateStartContainer(ctx, cli, imageName, config.CurrentDevTestSetupConfig.Postgres.ContainerName, hostConfig, cfg)
}

func StopPostgresContainer(ctx context.Context, cli *client.Client) error {
	return StopContainerByName(ctx, cli, config.CurrentDevTestSetupConfig.Postgres.ContainerName)
}
