package api

import (
	"DevTestSetup/internal/config"
	"DevTestSetup/internal/docker"
	"DevTestSetup/internal/entity"
	"context"
	"github.com/docker/docker/client"
	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type Server struct {
	App           *fiber.App
	DockerClient  *client.Client
	DockerContext context.Context
}

// NewContext initializes an AppContext
func NewServer(app *fiber.App, cli *client.Client, ctx context.Context) *Server {
	cli.NegotiateAPIVersion(ctx)
	return &Server{App: app, DockerClient: cli, DockerContext: ctx}
}

func (s *Server) Listen() {
	var wg sync.WaitGroup

	// Intercept Ctrl+C and SIGTERM
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	wg.Add(2)
	go s.StartRest(config.CurrentDevTestSetupConfig.Port, &wg)

	go func() {
		// Block until a signal is received.
		<-c
		log.Info("Gracefully shutting down...")
		s.Shutdown()
		os.Exit(0)
	}()

	wg.Wait()
}

func (s *Server) Setup() {
	err := docker.CreateNetworkAndVolumes(s.DockerContext, s.DockerClient)
	if err != nil {
		log.Fatal("could not create network or volume: ", err)
	}
}

func (s *Server) Autostart() {
	data := entity.Service{Nats: config.CurrentDevTestSetupConfig.Nats.Autostart,
		Redis:    config.CurrentDevTestSetupConfig.Redis.Autostart,
		Opa:      config.CurrentDevTestSetupConfig.Opa.Autostart,
		Postgres: config.CurrentDevTestSetupConfig.Postgres.Autostart,
		Hydra:    config.CurrentDevTestSetupConfig.Hydra.Autostart}
	if err := data.Start(s.DockerContext, s.DockerClient); err != nil {
		log.Fatal(err)
	}
}
