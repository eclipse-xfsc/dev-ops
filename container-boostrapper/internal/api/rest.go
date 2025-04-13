package api

import (
	"DevTestSetup/internal/config"
	"DevTestSetup/internal/entity"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"
	"strconv"
	"sync"
)

func (s *Server) GetServices(c *fiber.Ctx) error {
	nameFilter := filters.NewArgs()
	nameFilter.Add("name", config.CurrentDevTestSetupConfig.Nats.ContainerName)
	nameFilter.Add("name", config.CurrentDevTestSetupConfig.Redis.ContainerName)
	nameFilter.Add("name", config.CurrentDevTestSetupConfig.Opa.ContainerName)
	nameFilter.Add("name", config.CurrentDevTestSetupConfig.Postgres.ContainerName)
	nameFilter.Add("name", config.CurrentDevTestSetupConfig.Hydra.ContainerName)
	nameFilter.Add("name", config.CurrentDevTestSetupConfig.Hydra.Consent.ContainerName)
	nameFilter.Add("name", config.CurrentDevTestSetupConfig.Hydra.Migrate.ContainerName)

	containers, err := s.DockerClient.ContainerList(s.DockerContext, types.ContainerListOptions{All: true, Filters: nameFilter})
	if err != nil {
		err = fmt.Errorf("error occurred while getting list of containers: %w", err)
		return c.Status(fiber.StatusInternalServerError).JSON(err.Error())
	}

	containerInfos := make([]entity.ContainerInfo, 0)
	for _, cont := range containers {
		for _, name := range cont.Names {
			containerInfo := entity.ContainerInfo{Name: name, ID: cont.ID, Image: cont.Image, Status: cont.Status}
			containerInfos = append(containerInfos, containerInfo)
		}
	}

	return c.JSON(containerInfos)
}

func (s *Server) StartServices(c *fiber.Ctx) error {
	data := new(entity.Service)

	if err := c.BodyParser(data); err != nil {
		err = fmt.Errorf("error occurred while parsing request body: %w", err)
		return c.Status(fiber.StatusBadRequest).JSON(err.Error())
	}

	if err := data.Start(s.DockerContext, s.DockerClient); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(err.Error())
	}

	return c.SendStatus(fiber.StatusCreated)
}

func (s *Server) StopServices(c *fiber.Ctx) error {
	data := new(entity.Service)

	if err := c.BodyParser(data); err != nil {
		err = fmt.Errorf("error occurred while parsing request body: %w", err)
		return c.Status(fiber.StatusBadRequest).JSON(err.Error())
	}

	if err := data.Stop(s.DockerContext, s.DockerClient); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(err.Error())
	}

	return c.SendStatus(fiber.StatusAccepted)
}

func (s *Server) DefineEndpoints() {
	s.App.Get("/v1/services", s.GetServices)
	s.App.Post("/v1/services", s.StartServices)
	s.App.Delete("/v1/services", s.StopServices)
}

func (s *Server) StartRest(port int, wg *sync.WaitGroup) {
	s.App = fiber.New()

	log.Info("start serving rest endpoints!")
	defer wg.Done()

	s.DefineEndpoints()

	err := s.App.Listen(":" + strconv.Itoa(port))
	if err != nil {
		panic(err)
	}
}

func (s *Server) Shutdown() {
	if err := s.App.Shutdown(); err != nil {
		log.Fatal("Error shutting down server:", err)
	}

	data := entity.Service{true, true, true, true, true}
	if err := data.Stop(s.DockerContext, s.DockerClient); err != nil {
		log.Fatal(err)
	}

	log.Info("All containers stopped.")
	log.Info("Server stopped")
}
