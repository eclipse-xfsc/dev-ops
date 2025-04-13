package cmd

import (
	"DevTestSetup/internal/api"
	"DevTestSetup/internal/config"
	"context"
	"github.com/docker/docker/client"
	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var configFilePath string

var RootCmd = &cobra.Command{
	Use:   "container-bootstrapper",
	Short: "A service for selectively starting containers for testing",
	Run: func(cmd *cobra.Command, args []string) {
		err := config.LoadConfig(configFilePath)
		if err != nil {
			log.Fatal(err)
		}

		// create a new docker client
		cli, err := client.NewClientWithOpts(client.WithAPIVersionNegotiation(), client.FromEnv)
		if err != nil {
			log.Fatal(err)
		}

		s := api.NewServer(fiber.New(), cli, context.Background())

		s.Setup()
		s.Autostart()
		s.Listen()
	},
}

func init() {
	// here we create a --config flag that will take a string argument
	RootCmd.PersistentFlags().StringVar(&configFilePath, "config", "./config.yaml", "config file path")
}
