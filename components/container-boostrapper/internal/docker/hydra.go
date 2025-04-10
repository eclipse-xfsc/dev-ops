package docker

import (
	"DevTestSetup/internal/config"
	"context"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

func StartHydraContainer(ctx context.Context, cli *client.Client) (string, error) {
	err := RemoveExistingContainer(ctx, cli, config.CurrentDevTestSetupConfig.Hydra.ContainerName)
	if err != nil {
		return "", nil
	}

	imageName := config.CurrentDevTestSetupConfig.Hydra.Image + ":" + config.CurrentDevTestSetupConfig.Hydra.Tag

	// Define host networking configuration.
	hostConfig := &container.HostConfig{
		PortBindings: nat.PortMap{
			"4444/tcp": []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: config.CurrentDevTestSetupConfig.Hydra.PublicPort,
				},
			},
			"4445/tcp": []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: config.CurrentDevTestSetupConfig.Hydra.AdminPort,
				},
			},
			"5555/tcp": []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: config.CurrentDevTestSetupConfig.Hydra.TokenPort,
				},
			},
		},
		NetworkMode: "intranet",
	}

	// Define container configuration
	cfg := &container.Config{
		Image: imageName,
		Cmd: []string{
			"serve",
			"all",
			"--dev",
		},
		Env: []string{
			"URLS_SELF_ISSUER=http://127.0.0.1:" + config.CurrentDevTestSetupConfig.Hydra.PublicPort,
			"URLS_CONSENT=http://127.0.0.1:" + config.CurrentDevTestSetupConfig.Hydra.Consent.Port + "/consent",
			"URLS_LOGIN=http://127.0.0.1:" + config.CurrentDevTestSetupConfig.Hydra.Consent.Port + "/login",
			"URLS_LOGOUT=http://127.0.0.1:" + config.CurrentDevTestSetupConfig.Hydra.Consent.Port + "/logout",
			"SECRETS_SYSTEM=youReallyNeedToChangeThis",
			"OIDC_SUBJECT_TYPES_SUPPORTED=public,pairwise",
			"OIDC_SUBJECT_TYPE_PAIRWISE_SALT=youReallyNeedToChangeThis",
			"SERVE_COOKIES_SAME_SITE_MODE=Lax",
			"DSN=postgres://" + config.CurrentDevTestSetupConfig.Postgres.User + ":" +
				config.CurrentDevTestSetupConfig.Postgres.Password + "@" +
				config.CurrentDevTestSetupConfig.Postgres.ContainerName + ":" +
				config.CurrentDevTestSetupConfig.Postgres.Port + "/" +
				config.CurrentDevTestSetupConfig.Postgres.Db +
				"?sslmode=disable&max_conns=20&max_idle_conns=4",
		},
		ExposedPorts: nat.PortSet{
			nat.Port(config.CurrentDevTestSetupConfig.Hydra.PublicPort + "/tcp"): struct{}{},
			nat.Port(config.CurrentDevTestSetupConfig.Hydra.AdminPort + "/tcp"):  struct{}{},
			nat.Port(config.CurrentDevTestSetupConfig.Hydra.TokenPort + "/tcp"):  struct{}{},
		},
	}

	return PullCreateStartContainer(ctx, cli, imageName, config.CurrentDevTestSetupConfig.Hydra.ContainerName, hostConfig, cfg)
}

func StopHydraContainer(ctx context.Context, cli *client.Client) error {
	return StopContainerByName(ctx, cli, config.CurrentDevTestSetupConfig.Hydra.ContainerName)
}

func StartHydraMigrateContainer(ctx context.Context, cli *client.Client) (string, error) {
	err := RemoveExistingContainer(ctx, cli, config.CurrentDevTestSetupConfig.Hydra.Migrate.ContainerName)
	if err != nil {
		return "", nil
	}

	imageName := config.CurrentDevTestSetupConfig.Hydra.Migrate.Image + ":" + config.CurrentDevTestSetupConfig.Hydra.Migrate.Tag

	// Define host networking configuration.
	hostConfig := &container.HostConfig{
		NetworkMode: "intranet",
	}

	// Define container configuration
	cfg := &container.Config{
		Image: imageName,
		Cmd: []string{
			"migrate",
			"sql",
			"-e",
			"--yes",
		},
		Env: []string{
			"URLS_SELF_ISSUER=http://127.0.0.1:" + config.CurrentDevTestSetupConfig.Hydra.PublicPort,
			"URLS_CONSENT=http://127.0.0.1:" + config.CurrentDevTestSetupConfig.Hydra.Consent.Port + "/consent",
			"URLS_LOGIN=http://127.0.0.1:" + config.CurrentDevTestSetupConfig.Hydra.Consent.Port + "/login",
			"URLS_LOGOUT=http://127.0.0.1:" + config.CurrentDevTestSetupConfig.Hydra.Consent.Port + "/logout",
			"SECRETS_SYSTEM=" + config.CurrentDevTestSetupConfig.Hydra.SecretsSystem,
			"OIDC_SUBJECT_TYPES_SUPPORTED=public,pairwise",
			"OIDC_SUBJECT_TYPE_PAIRWISE_SALT=" + config.CurrentDevTestSetupConfig.Hydra.OidcSubjectTypePairwiseSalt,
			"SERVE_COOKIES_SAME_SITE_MODE=Lax",
			"DSN=" + config.CurrentDevTestSetupConfig.Hydra.DataSourceName,
		},
	}

	return PullCreateStartContainer(ctx, cli, imageName, config.CurrentDevTestSetupConfig.Hydra.Migrate.ContainerName, hostConfig, cfg)
}

func StopHydraMigrateContainer(ctx context.Context, cli *client.Client) error {
	return StopContainerByName(ctx, cli, config.CurrentDevTestSetupConfig.Hydra.Migrate.ContainerName)
}

func StartHydraConsentContainer(ctx context.Context, cli *client.Client) (string, error) {
	err := RemoveExistingContainer(ctx, cli, config.CurrentDevTestSetupConfig.Hydra.Consent.ContainerName)
	if err != nil {
		return "", nil
	}

	imageName := config.CurrentDevTestSetupConfig.Hydra.Consent.Image + ":" + config.CurrentDevTestSetupConfig.Hydra.Consent.Tag

	// Define host networking configuration.
	hostConfig := &container.HostConfig{
		PortBindings: nat.PortMap{
			"3000/tcp": []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: config.CurrentDevTestSetupConfig.Hydra.Consent.Port,
				},
			},
		},
		NetworkMode: "intranet",
	}

	// Define container configuration
	cfg := &container.Config{
		Image: imageName,
		ExposedPorts: nat.PortSet{
			nat.Port(config.CurrentDevTestSetupConfig.Hydra.Consent.Port + "/tcp"): struct{}{},
		},
		Env: []string{
			"HYDRA_ADMIN_URL=http://hydra:" + config.CurrentDevTestSetupConfig.Hydra.AdminPort,
			"NODE_TLS_REJECT_UNAUTHORIZED=0",
			"PORT=" + config.CurrentDevTestSetupConfig.Hydra.Consent.Port,
		},
	}

	return PullCreateStartContainer(ctx, cli, imageName, config.CurrentDevTestSetupConfig.Hydra.Consent.ContainerName, hostConfig, cfg)
}

func StopHydraConsentContainer(ctx context.Context, cli *client.Client) error {
	return StopContainerByName(ctx, cli, config.CurrentDevTestSetupConfig.Hydra.Consent.ContainerName)
}
