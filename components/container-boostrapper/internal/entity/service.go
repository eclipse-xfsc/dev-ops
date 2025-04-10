package entity

import (
	"DevTestSetup/internal/docker"
	"context"
	"fmt"
	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
)

type Service struct {
	Nats     bool `json:"nats,omitempty"`
	Redis    bool `json:"redis,omitempty"`
	Opa      bool `json:"opa,omitempty"`
	Postgres bool `json:"postgres,omitempty"`
	Hydra    bool `json:"hydra,omitempty"`
}

func (s *Service) Start(ctx context.Context, cli *client.Client) error {
	start := map[*bool]func() error{
		&s.Nats: func() error {
			id, err := docker.StartNatsContainer(ctx, cli)
			if err != nil {
				err = fmt.Errorf("error occurred while starting Nats container: %w", err)
				return err
			}
			log.Infof("nats container started: %s", id)
			return nil
		},
		&s.Redis: func() error {
			id, err := docker.StartRedisContainer(ctx, cli)
			if err != nil {
				err = fmt.Errorf("error occurred while starting redis container: %w", err)
				return err
			}
			log.Infof("redis container started: %s", id)
			return nil
		},
		&s.Opa: func() error {
			id, err := docker.StartOpaContainer(ctx, cli)
			if err != nil {
				err = fmt.Errorf("error occurred while starting opa container: %w", err)
				return err
			}
			log.Infof("opa container started: %s", id)
			return nil
		},
		&s.Postgres: func() error {
			id, err := docker.StartPostgresContainer(ctx, cli)
			if err != nil {
				err = fmt.Errorf("error occurred while starting postgres container: %w", err)
				return err
			}
			log.Infof("postgres container started: %s", id)
			return nil
		},
		&s.Hydra: func() error {
			// Start hydra migrate first
			id, err := docker.StartHydraMigrateContainer(ctx, cli)
			if err != nil {
				err = fmt.Errorf("error occurred while starting hydra-migrate container: %w", err)
				return err
			}
			log.Infof("hydra-migrate container started: %s", id)

			// Start hydra consent next
			id, err = docker.StartHydraConsentContainer(ctx, cli)
			if err != nil {
				err = fmt.Errorf("error occurred while starting hydra-consent container: %w", err)
				return err
			}
			log.Infof("hydra-consent container started: %s", id)

			// Start hydra last
			id, err = docker.StartHydraContainer(ctx, cli)
			if err != nil {
				err = fmt.Errorf("error occurred while starting hydra container: %w", err)
				return err
			}
			log.Infof("hydra container started: %s", id)
			return nil
		},
	}

	for startCondition, startFunc := range start {
		if *startCondition {
			if err := startFunc(); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Service) Stop(ctx context.Context, cli *client.Client) error {
	stop := map[*bool]func() error{
		&s.Nats: func() error {
			err := docker.StopNatsContainer(ctx, cli)
			if err != nil {
				err = fmt.Errorf("error occurred while stopping Nats container: %w", err)
				return err
			}
			log.Infof("nats container stopped")
			return nil
		},
		&s.Redis: func() error {
			err := docker.StopRedisContainer(ctx, cli)
			if err != nil {
				err = fmt.Errorf("error occurred while stopping redis container: %w", err)
				return err
			}
			log.Infof("redis container stopped")
			return nil
		},
		&s.Opa: func() error {
			err := docker.StopOpaContainer(ctx, cli)
			if err != nil {
				err = fmt.Errorf("error occurred while stopping opa container: %w", err)
				return err
			}
			log.Infof("opa container stopped")
			return nil
		},
		&s.Postgres: func() error {
			err := docker.StopPostgresContainer(ctx, cli)
			if err != nil {
				err = fmt.Errorf("error occurred while stopping postgres container: %w", err)
				return err
			}
			log.Infof("postgres container stopped")
			return nil
		},
		&s.Hydra: func() error {
			// Start hydra migrate first
			err := docker.StopHydraMigrateContainer(ctx, cli)
			if err != nil {
				err = fmt.Errorf("error occurred while stopping hydra-migrate container: %w", err)
				return err
			}
			log.Infof("hydra-migrate container stopped")

			// Start hydra consent next
			err = docker.StopHydraConsentContainer(ctx, cli)
			if err != nil {
				err = fmt.Errorf("error occurred while stopping hydra-consent container: %w", err)
				return err
			}
			log.Infof("hydra-consent container stopped")

			// Start hydra last
			err = docker.StopHydraContainer(ctx, cli)
			if err != nil {
				err = fmt.Errorf("error occurred while stopping hydra container: %w", err)
				return err
			}
			log.Infof("hydra container stopped")
			return nil
		},
	}

	for stopCondition, stopFunc := range stop {
		if *stopCondition {
			if err := stopFunc(); err != nil {
				return err
			}
		}
	}
	return nil
}
