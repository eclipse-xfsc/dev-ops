package config

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"strings"
)

type containerBootstrapperSetupConfig struct {
	LogLevel string `mapstructure:"logLevel"`
	Port     int    `mapstructure:"port"`
	Docker   struct {
		RepoUrl  string `mapstructure:"repoUrl"`
		Username string `mapstructure:"username"`
		Password string `mapstructure:"password"`
	}
	Nats struct {
		Image          string `mapstructure:"image"`
		Tag            string `mapstructure:"tag"`
		ContainerName  string `mapstrucute:"containerName"`
		ClusterName    string `mapstructure:"clusterName"`
		ServerPort     string `mapstructure:"serverPort"`
		MonitoringPort string `mapstructure:"monitoringPort"`
		Autostart      bool
	} `mapstructure:"nats"`
	Redis struct {
		Image         string `mapstructure:"image"`
		Tag           string `mapstructure:"tag"`
		ContainerName string `mapstrucute:"containerName"`
		Port          string `mapstructure:"port"`
		Autostart     bool
	} `mapstructure:"redis"`
	Opa struct {
		Image         string `mapstructure:"image"`
		Tag           string `mapstructure:"tag"`
		ContainerName string `mapstrucute:"containerName"`
		Port          string `mapstructure:"port"`
		Autostart     bool
	} `mapstructure:"opa"`
	Postgres struct {
		Image         string `mapstructure:"image"`
		Tag           string `mapstructure:"tag"`
		ContainerName string `mapstrucute:"containerName"`
		Port          string `mapstructure:"port"`
		User          string `mapstructure:"user"`
		Password      string `mapstructure:"password"`
		Db            string `mapstructure:"db"`
		Autostart     bool
	} `mapstructure:"postgres"`
	Hydra struct {
		Image                       string `mapstructure:"image"`
		Tag                         string `mapstructure:"tag"`
		ContainerName               string `mapstrucute:"containerName"`
		PublicPort                  string `mapstructure:"publicPort"`
		AdminPort                   string `mapstructure:"adminPort"`
		TokenPort                   string `mapstructure:"tokenPort"`
		DataSourceName              string `mapstructure:"dsn"`
		SecretsSystem               string `mapstructure:"secretsSystem"`
		OidcSubjectTypePairwiseSalt string `mapstructure:"pairwiseSalt"`
		Autostart                   bool
		Migrate                     struct {
			Image         string `mapstructure:"image"`
			Tag           string `mapstructure:"tag"`
			ContainerName string `mapstrucute:"containerName"`
		} `mapstructure:"migrate"`
		Consent struct {
			Image         string `mapstructure:"image"`
			Tag           string `mapstructure:"tag"`
			ContainerName string `mapstrucute:"containerName"`
			Port          string `mapstructure:"port"`
		} `mapstructure:"consent"`
	} `mapstructure:"hydra"`
}

var CurrentDevTestSetupConfig containerBootstrapperSetupConfig

func LoadConfig(configFilePath string) error {

	if !fileExists(configFilePath) {
		log.Warningf("file %s does not exist, setting default values.", configFilePath)
	} else {
		log.Infof("Loading config from file: %s", configFilePath)
		readConfig(configFilePath)
	}

	setDefaults()

	if err := viper.Unmarshal(&CurrentDevTestSetupConfig); err != nil {
		return err
	}
	setLogLevel()
	return nil
}

func readConfig(configFilePath string) {
	viper.SetConfigFile(configFilePath)

	viper.SetEnvPrefix("CONTAINER_BOOTSTRAPPER")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			log.Warning("Configuration not found but environment variables will be taken into account.")
		}
	}

	containersMap := map[string]*bool{
		"nats":     &CurrentDevTestSetupConfig.Nats.Autostart,
		"redis":    &CurrentDevTestSetupConfig.Redis.Autostart,
		"opa":      &CurrentDevTestSetupConfig.Opa.Autostart,
		"postgres": &CurrentDevTestSetupConfig.Postgres.Autostart,
		"hydra":    &CurrentDevTestSetupConfig.Hydra.Autostart,
	}

	for key, _ := range viper.AllSettings() {
		if autostart, ok := containersMap[key]; ok {
			log.Infof("autostarting %s container", key)
			*autostart = true
		}
	}

	viper.AutomaticEnv()
}

func setDefaults() {
	viper.SetDefault("logLevel", "info")
	viper.SetDefault("port", 3131)

	// docker config
	viper.SetDefault("docker.repoUrl", "")
	viper.SetDefault("docker.username", "")
	viper.SetDefault("docker.password", "")

	// nats config
	viper.SetDefault("nats.image", "nats")
	viper.SetDefault("nats.tag", "latest")
	viper.SetDefault("nats.containerName", "nats_dev")
	viper.SetDefault("nats.clusterName", "NATS")
	viper.SetDefault("nats.serverPort", "4222")
	viper.SetDefault("nats.monitoringPort", "8222")

	// redis config
	viper.SetDefault("redis.image", "redis")
	viper.SetDefault("redis.tag", "alpine")
	viper.SetDefault("redis.containerName", "redis_dev")
	viper.SetDefault("redis.port", "6379")

	// opa config
	viper.SetDefault("opa.image", "openpolicyagent/opa")
	viper.SetDefault("opa.tag", "latest")
	viper.SetDefault("opa.containerName", "opa_dev")
	viper.SetDefault("opa.port", "8181")

	// postgres config
	viper.SetDefault("postgres.image", "postgres")
	viper.SetDefault("postgres.tag", "alpine")
	viper.SetDefault("postgres.containerName", "postgresd_dev")
	viper.SetDefault("postgres.port", "5432")
	viper.SetDefault("postgres.user", "hydra")
	viper.SetDefault("postgres.password", "dev")
	viper.SetDefault("postgres.db", "hydra")

	// hydra config
	viper.SetDefault("hydra.image", "oryd/hydra")
	viper.SetDefault("hydra.tag", "v2.2.0-rc.3")
	viper.SetDefault("hydra.containerName", "hydra_dev")
	viper.SetDefault("hydra.publicPort", "4444")
	viper.SetDefault("hydra.adminPort", "4445")
	viper.SetDefault("hydra.tokenPort", "5555")
	viper.SetDefault("hydra.dsn", "postgres://hydra:dev@postgresd_dev:5432/hydra"+
		"?sslmode=disable&max_conns=20&max_idle_conns=4")
	viper.SetDefault("hydra.secretsSystem", "youReallyNeedToChangeThis")
	viper.SetDefault("hydra.pairwiseSalt", "youReallyNeedToChangeThis")
	viper.SetDefault("hydra.migrate.image", "oryd/hydra")
	viper.SetDefault("hydra.migrate.tag", "v2.2.0-rc.3")
	viper.SetDefault("hydra.migrate.containerName", "hydra-migrate_dev")
	viper.SetDefault("hydra.consent.image", "oryd/hydra-login-consent-node")
	viper.SetDefault("hydra.consent.tag", "v2.2.0-rc.3")
	viper.SetDefault("hydra.consent.containerName", "hydra-consent_dev")
	viper.SetDefault("hydra.consent.port", "3000")
}

func setLogLevel() {
	var logLevel log.Level
	switch strings.ToLower(CurrentDevTestSetupConfig.LogLevel) {
	case "trace":
		logLevel = log.TraceLevel
	case "debug":
		logLevel = log.DebugLevel
	case "info":
		logLevel = log.InfoLevel
	case "error":
		logLevel = log.ErrorLevel
	case "fatal":
		logLevel = log.FatalLevel
	case "panic":
		logLevel = log.PanicLevel
	default:
		logLevel = log.WarnLevel
	}
	log.SetLevel(logLevel)
	log.Infof("loglevel set to %s", logLevel.String())
}

func fileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !errors.Is(err, os.ErrNotExist)
}
