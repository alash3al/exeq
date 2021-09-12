package config

import (
	"fmt"

	"github.com/hashicorp/hcl/v2/hclsimple"
)

type Config struct {
	Queue      *QueueConfig      `hcl:"queue,block"`
	HTTPServer *HTTPServerConfig `hcl:"http_server,block"`
	Macros     []*MacroConfig    `hcl:"macro,block"`

	macrosMap map[string]*MacroConfig
}

func BootFromFile(filename string) (config Config, err error) {
	err = hclsimple.DecodeFile(filename, nil, &config)

	if err != nil {
		return
	}

	config.macrosMap = map[string]*MacroConfig{}

	err = config.validateCommands()

	return
}

func (config *Config) LookupMacro(name string) (m *MacroConfig, ok bool) {
	m, ok = config.macrosMap[name]

	return
}

func (config *Config) SetupMounts() error {
	for _, macro := range config.Macros {
		if err := macro.setupMounts(); err != nil {
			return err
		}
	}

	return nil
}

func (config *Config) validateCommands() error {
	for _, macro := range config.Macros {
		if _, ok := config.macrosMap[macro.Name]; ok {
			return fmt.Errorf("duplicate macro (%s)", macro.Name)
		}

		config.macrosMap[macro.Name] = macro

		if err := macro.validateCommand(); err != nil {
			return err
		}
	}

	return nil
}
