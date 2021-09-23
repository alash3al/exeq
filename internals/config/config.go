package config

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/hashicorp/hcl/v2/hclsimple"
)

type Config struct {
	LogLevel   string            `hcl:"log_level"`
	Queue      *QueueConfig      `hcl:"queue,block"`
	HTTPServer *HTTPServerConfig `hcl:"http_server,block"`
	Macros     []*MacroConfig    `hcl:"macro,block"`

	macrosMap map[string]*MacroConfig
}

func BootFromFile(filename string) (*Config, error) {
	var config Config

	fileContent, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	fileContent = []byte(os.ExpandEnv(string(fileContent)))

	if err := hclsimple.Decode(filename, fileContent, nil, &config); err != nil {
		return nil, err
	}

	config.macrosMap = map[string]*MacroConfig{}

	err = config.validateCommands()

	return &config, err
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
