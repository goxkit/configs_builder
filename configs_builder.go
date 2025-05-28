// Copyright (c) 2025, The GoKit Authors
// MIT License
// All rights reserved.

// Package configsbuilder provides a fluent interface for building application configurations.
// It simplifies the process of loading configurations from environment variables and .env files
// for various components of an application such as HTTP, messaging, databases, etc.
package configsbuilder

import (
	"os"
	"reflect"

	"github.com/goxkit/configs"
	noopLogging "github.com/goxkit/logging/noop"
	otlpLogging "github.com/goxkit/logging/otlp"
	otlpTracing "github.com/goxkit/tracing/otlp"
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
	cfgs, err := b.newConfigs()
	if err != nil {
		return nil, err
	}

	if err := b.setupObservability(cfgs); err != nil {
		return nil, err
	}

	// Load component-specific configurations based on what was enabled
	if b.http {
		cfgs.HTTPConfigs = &configs.HTTPConfigs{}
		b.loadStructDefaults(cfgs.Custom, cfgs.HTTPConfigs)
		err = cfgs.Custom.Unmarshal(cfgs.HTTPConfigs)
		if err != nil {
			cfgs.Logger.Error("failed to unmarshal HTTP configs", zap.Error(err))
			return nil, err
		}
	}

	if b.postgres {
		cfgs.PostgresConfigs = &configs.PostgresConfigs{}
		b.loadStructDefaults(cfgs.Custom, cfgs.PostgresConfigs)
		err = cfgs.Custom.Unmarshal(cfgs.PostgresConfigs)
		if err != nil {
			cfgs.Logger.Error("failed to unmarshal Postgres configs", zap.Error(err))
			return nil, err
		}
	}

	if b.identity {
		cfgs.IdentityConfigs = &configs.IdentityConfigs{}
		b.loadStructDefaults(cfgs.Custom, cfgs.IdentityConfigs)
		err = cfgs.Custom.Unmarshal(cfgs.IdentityConfigs)
		if err != nil {
			cfgs.Logger.Error("failed to unmarshal Identity configs", zap.Error(err))
			return nil, err
		}
	}

	if b.mqtt {
		cfgs.MQTTConfigs = &configs.MQTTConfigs{}
		b.loadStructDefaults(cfgs.Custom, cfgs.MQTTConfigs)
		err = cfgs.Custom.Unmarshal(cfgs.MQTTConfigs)
		if err != nil {
			cfgs.Logger.Error("failed to unmarshal MQTT configs", zap.Error(err))
			return nil, err
		}
	}

	if b.rabbitmq {
		cfgs.RabbitMQConfigs = &configs.RabbitMQConfigs{}
		b.loadStructDefaults(cfgs.Custom, cfgs.RabbitMQConfigs)
		err = cfgs.Custom.Unmarshal(cfgs.RabbitMQConfigs)
		if err != nil {
			cfgs.Logger.Error("failed to unmarshal RabbitMQ configs", zap.Error(err))
			return nil, err
		}
	}

	if b.aws {
		cfgs.AWSConfigs = &configs.AWSConfigs{}
		b.loadStructDefaults(cfgs.Custom, cfgs.AWSConfigs)
		err = cfgs.Custom.Unmarshal(cfgs.AWSConfigs)
		if err != nil {
			cfgs.Logger.Error("failed to unmarshal AWS configs", zap.Error(err))
			return nil, err
		}
	}

	if b.dynamoDB {
		cfgs.DynamoDBConfigs = &configs.DynamoDBConfigs{}
		b.loadStructDefaults(cfgs.Custom, cfgs.DynamoDBConfigs)
		err = cfgs.Custom.Unmarshal(cfgs.DynamoDBConfigs)
		if err != nil {
			cfgs.Logger.Error("failed to unmarshal Dynamo configs", zap.Error(err))
			return nil, err
		}
	}

	return cfgs, nil
}

func (b *configsBuilder) newConfigs() (*configs.Configs, error) {
	v, err := b.setupViper()
	if err != nil {
		return nil, err
	}

	appConfigs := &configs.AppConfigs{}
	b.loadStructDefaults(v, appConfigs)
	err = v.Unmarshal(appConfigs)
	if err != nil {
		return nil, err
	}

	otlpConfigs := &configs.OTLPConfigs{}
	b.loadStructDefaults(v, otlpConfigs)
	err = v.Unmarshal(otlpConfigs)
	if err != nil {
		return nil, err
	}

	return &configs.Configs{
		AppConfigs:  appConfigs,
		OTLPConfigs: otlpConfigs,
		Custom:      v,
	}, nil
}

func (c *configsBuilder) setupViper() (*viper.Viper, error) {
	env := configs.NewEnvironment(os.Getenv("GO_ENV"))

	v := viper.New()
	v.SetConfigFile(env.EnvFile())
	v.SetConfigType("env")
	v.AutomaticEnv()
	err := v.ReadInConfig()
	if err != nil {
		return nil, err
	}

	return v, nil
}

func (b *configsBuilder) setupObservability(cfgs *configs.Configs) error {
	var err error

	if b.otlp {
		cfgs.Logger, err = otlpLogging.Install(cfgs)
		if err != nil {
			cfgs.Logger.Error("failed to install OTLP logger", zap.Error(err))
			return err
		}

		_, err = otlpTracing.Install(cfgs)
		if err != nil {
			cfgs.Logger.Error("failed to install OTLP tracing", zap.Error(err))
			return err
		}

		// _, err = otlpMetrics.Install(&cfgs)
		// if err != nil {
		// 	return nil, err
		// }

		return nil
	}

	cfgs.Logger, err = noopLogging.Install(cfgs)
	if err != nil {
		cfgs.Logger.Error("failed to install Noop logger", zap.Error(err))
		return err
	}

	return nil
}

// loadStructDefaults takes a struct and loads default values defined in 'default' tags
// into the specified Viper instance. This allows setting defaults directly in struct tags.
func (c *configsBuilder) loadStructDefaults(v *viper.Viper, structPtr interface{}) {
	// Get the reflect Value and Type of the struct
	val := reflect.ValueOf(structPtr)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// Must be a struct
	if val.Kind() != reflect.Struct {
		return
	}

	typ := val.Type()

	// Iterate through all fields in the struct
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)

		// Get the default tag if it exists
		defaultVal, ok := field.Tag.Lookup("default")
		if !ok || defaultVal == "" {
			continue
		}

		// Get the mapstructure tag which defines how viper maps the environment variable
		envKey, ok := field.Tag.Lookup("mapstructure")
		if !ok || envKey == "" {
			continue
		}

		// Set the default value in Viper
		v.SetDefault(envKey, defaultVal)
	}
}
