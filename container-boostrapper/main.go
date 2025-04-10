package main

import (
	"DevTestSetup/internal/cmd"
	log "github.com/sirupsen/logrus"
	"os"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		log.Error("Error: ", err)
		os.Exit(1)
	}
}
