package main

import (
	"net/http"
	"os"

	"github.com/pivotal-cf/brokerapi"
	"github.com/pivotal-golang/lager"

	cf_lager "github.com/cloudfoundry-incubator/cf-lager"
	"github.com/pivotal-cf/cf-redis-broker/availability"
	"github.com/pivotal-cf/cf-redis-broker/broker"
	"github.com/pivotal-cf/cf-redis-broker/brokerconfig"
	"github.com/pivotal-cf/cf-redis-broker/debug"
	"github.com/pivotal-cf/cf-redis-broker/process"
	"github.com/pivotal-cf/cf-redis-broker/redis"
	"github.com/pivotal-cf/cf-redis-broker/system"
)

func main() {
	brokerConfigPath := configPath()
	brokerLogger := cf_lager.New("redis-broker")
	brokerLogger.Info("Config File: " + brokerConfigPath)

	config, err := brokerconfig.ParseConfig(brokerConfigPath)
	if err != nil {
		brokerLogger.Fatal("Loading config file", err, lager.Data{
			"broker-config-path": brokerConfigPath,
		})
	}

	commandRunner := system.OSCommandRunner{
		Logger: brokerLogger,
	}

	localRepo := &redis.LocalRepository{
		RedisConf: config.RedisConfiguration,
	}

	processController := &redis.OSProcessController{
		CommandRunner:            commandRunner,
		InstanceInformer:         localRepo,
		Logger:                   brokerLogger,
		ProcessChecker:           &process.ProcessChecker{},
		ProcessKiller:            &process.ProcessKiller{},
		WaitUntilConnectableFunc: availability.Check,
	}

	localCreator := &redis.LocalInstanceCreator{
		FindFreePort:            system.FindFreePort,
		RedisConfiguration:      config.RedisConfiguration,
		ProcessController:       processController,
		CredentialsGenerator:    redis.RandomCredentialsGenerator{},
		LocalInstanceRepository: localRepo,
	}

	remoteRepo, err := redis.NewRemoteRepository(&config)
	if err != nil {
		brokerLogger.Fatal("Error initializing remote repository", err)
	}

	serviceBroker := &broker.RedisServiceBroker{
		InstanceCreators: map[string]broker.InstanceCreator{
			"shared":    localCreator,
			"dedicated": remoteRepo,
		},
		InstanceBinders: map[string]broker.InstanceBinder{
			"shared":    localRepo,
			"dedicated": remoteRepo,
		},
		Config: config,
	}

	debugHandler := debug.NewHandler(remoteRepo)
	debugHandler = debug.BuildAuthenticatedHandler(
		config.AuthConfiguration.Username,
		config.AuthConfiguration.Password,
		debugHandler,
	)

	brokerCredentials := brokerapi.BrokerCredentials{
		Username: config.AuthConfiguration.Username,
		Password: config.AuthConfiguration.Password,
	}
	brokerAPI := brokerapi.New(serviceBroker, brokerLogger, brokerCredentials)
	http.HandleFunc("/debug", debugHandler)
	http.Handle("/", brokerAPI)

	if http.ListenAndServe(":"+config.Port, nil) != nil {
		brokerLogger.Error("http-listen", err)
	}
}

func configPath() string {
	brokerConfigYamlPath := os.Getenv("BROKER_CONFIG_PATH")
	if brokerConfigYamlPath == "" {
		panic("BROKER_CONFIG_PATH not set")
	}
	return brokerConfigYamlPath
}