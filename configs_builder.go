// Copyright (c) 2025, The GoKit Authors
// MIT License
// All rights reserved.

// Package configsbuilder provides a fluent interface for building application configurations.
// It simplifies the process of loading configurations from environment variables and .env files
// for various components of an application such as HTTP, messaging, databases, etc.
package configsbuilder

import (
	"os"

	"github.com/goxkit/configs"
	"github.com/goxkit/logging"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type (
	// ConfigsBuilder defines the interface for the builder pattern used to construct configurations.
	// It provides methods to specify which configuration components should be included.
	ConfigsBuilder interface {
		// HTTP enables HTTP server configuration loading
		HTTP() ConfigsBuilder
		// Opentelemetry OTLP enables OpenTelemetry logging, tracing and metrics configuration
		Otlp() ConfigsBuilder
		// Postgres enables SQL database configuration loading
		Postgres() ConfigsBuilder
		// Identity enables identity/authentication configuration loading
		Identity() ConfigsBuilder
		// MQTT enables MQTT client configuration loading
		MQTT() ConfigsBuilder
		// RabbitMQ enables RabbitMQ configuration loading
		RabbitMQ() ConfigsBuilder
		// AWS enables AWS configuration loading
		AWS() ConfigsBuilder
		// DynamoDB enables DynamoDB configuration loading
		DynamoDB() ConfigsBuilder
		// Build processes all enabled configurations and returns the complete config object
		Build() (*configs.Configs, error)
	}

	// configsBuilder implements the ConfigsBuilder interface and tracks which configurations to load
	configsBuilder struct {
		Err error

		http     bool
		otlp     bool
		postgres bool
		identity bool
		mqtt     bool
		rabbitmq bool
		aws      bool
		dynamoDB bool
	}
)

// NewConfigsBuilder creates a new instance of ConfigsBuilder with no configurations enabled
func NewConfigsBuilder() ConfigsBuilder {
	return &configsBuilder{}
}

// HTTP enables HTTP configuration loading in the builder
func (b *configsBuilder) HTTP() ConfigsBuilder {
	b.http = true
	return b
}

// Tracing enables tracing configuration loading in the builder
func (b *configsBuilder) Otlp() ConfigsBuilder {
	b.otlp = true
	return b
}

// SQLDatabase enables SQL database configuration loading in the builder
func (b *configsBuilder) Postgres() ConfigsBuilder {
	b.postgres = true
	return b
}

// Identity enables identity configuration loading in the builder
func (b *configsBuilder) Identity() ConfigsBuilder {
	b.identity = true
	return b
}

// MQTT enables MQTT configuration loading in the builder
func (b *configsBuilder) MQTT() ConfigsBuilder {
	b.mqtt = true
	return b
}

// RabbitMQ enables RabbitMQ configuration loading in the builder
func (b *configsBuilder) RabbitMQ() ConfigsBuilder {
	b.rabbitmq = true
	return b
}

// AWS enables AWS configuration loading in the builder
func (b *configsBuilder) AWS() ConfigsBuilder {
	b.aws = true
	return b
}

// DynamoDB enables DynamoDB configuration loading in the builder
func (b *configsBuilder) DynamoDB() ConfigsBuilder {
	b.dynamoDB = true
	return b
}

// Build processes all enabled configurations and returns the complete configs object.
// It reads environment variables, loads .env files, and constructs the configuration
// based on the enabled features. Returns an error if any configuration fails to load.
func (b *configsBuilder) Build() (*configs.Configs, error) {
	env := configs.NewEnvironment(os.Getenv("GO_ENV"))

	v := viper.New()
	v.SetConfigFile(env.EnvFile())
	v.SetConfigType("env")
	v.AutomaticEnv()
	err := v.ReadInConfig()
	if err != nil {
		return nil, err
	}

	cfgs := configs.Configs{Custom: v}

	// Load application base configurations
	err = v.Unmarshal(cfgs.AppConfigs)
	if err != nil {
		return nil, err
	}

	// Initialize the logger
	logger, err := logging.NewDefaultLogger(&cfgs)
	if err != nil {
		return nil, err
	}

	cfgs.Logger = logger.(*zap.Logger)

	// Load component-specific configurations based on what was enabled
	if b.http {
		err = v.Unmarshal(cfgs.HTTPConfigs)
		if err != nil {
			logger.Error("failed to unmarshal HTTP configs", zap.Error(err))
			return nil, err
		}
	}

	if b.otlp {
		err = v.Unmarshal(cfgs.OTLPConfigs)
		if err != nil {
			logger.Error("failed to unmarshal OTLP configs", zap.Error(err))
			return nil, err
		}
	}

	if b.postgres {
		err = v.Unmarshal(cfgs.PostgresConfigs)
		if err != nil {
			logger.Error("failed to unmarshal Postgres configs", zap.Error(err))
			return nil, err
		}
	}

	if b.identity {
		err = v.Unmarshal(cfgs.IdentityConfigs)
		if err != nil {
			logger.Error("failed to unmarshal Identity configs", zap.Error(err))
			return nil, err
		}
	}

	if b.mqtt {
		err = v.Unmarshal(cfgs.MQTTConfigs)
		if err != nil {
			logger.Error("failed to unmarshal MQTT configs", zap.Error(err))
			return nil, err
		}
	}

	if b.rabbitmq {
		err = v.Unmarshal(cfgs.RabbitMQConfigs)
		if err != nil {
			logger.Error("failed to unmarshal RabbitMQ configs", zap.Error(err))
			return nil, err
		}
	}

	if b.aws {
		err = v.Unmarshal(cfgs.AWSConfigs)
		if err != nil {
			logger.Error("failed to unmarshal AWS configs", zap.Error(err))
			return nil, err
		}
	}

	if b.dynamoDB {
		err = v.Unmarshal(cfgs.DynamoDBConfigs)
		if err != nil {
			logger.Error("failed to unmarshal Dynamo configs", zap.Error(err))
			return nil, err
		}
	}

	return &cfgs, nil
}
