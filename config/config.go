package config

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2/hclsimple"
)

type Config struct {
	Queue      QueueConfig      `hcl:"queue,block"`
	HTTPServer HTTPServerConfig `hcl:"http_server,block"`
	// Macros     []MacroConfig    `hcl:"macro,block"`
}

type QueueConfig struct {
	DSN           string `hcl:"dsn"`
	WorkersCount  int    `hcl:"workers_count"`
	PollDuration  string `hcl:"poll_duration"`
	RetryAttempts int    `hcl:"retry_attempts"`
}

type HTTPServerConfig struct {
	ListenAddr string `hcl:"listen"`
}

type MacroConfig struct {
	Name    string `hcl:"name,label"`
	Command string `hcl:"command"`
	Mounts  []struct {
		Filename string `hcl:"filename,label"`
		Content  string `hcl:"content"`
	} `hcl:"mount,block"`
}

func (macro *MacroConfig) ValidateCommand() error {
	cmdParts := strings.Split(strings.TrimSpace(macro.Command), " ")

	if len(cmdParts) < 1 {
		return fmt.Errorf("there is no command for the macro (%s)", macro.Name)
	}

	if _, err := exec.LookPath(cmdParts[0]); err != nil {
		return err
	}

	return nil
}

func (macro *MacroConfig) CreateMounts() (err error) {
	if len(macro.Mounts) < 1 {
		return
	}

	for _, mount := range macro.Mounts {
		if mount.Filename == "" {
			return fmt.Errorf("the macro (%s) has a macro that doesn't have a filename value", macro.Name)
		}

		dir := filepath.Dir(mount.Filename)

		// TODO Should we make the mounted file os.FileMode configurable?
		osFileMode := 0755

		if err := os.MkdirAll(dir, fs.FileMode(osFileMode)); err != nil {
			return fmt.Errorf(
				"cannot create the mount (%s) of the macro (%s) due to the following error (%s)",
				mount.Filename,
				macro.Name,
				err.Error(),
			)
		}

		if err := ioutil.WriteFile(mount.Filename, []byte(mount.Content), fs.FileMode(osFileMode)); err != nil {
			return fmt.Errorf(
				"cannot write the content of the mount (%s) of the macro (%s) due to the following error (%s)",
				mount.Filename,
				macro.Name,
				err.Error(),
			)
		}
	}

	return
}

func LoadFromFile(filename string) (config Config, err error) {
	err = hclsimple.DecodeFile(filename, nil, &config)

	if err != nil {
		return
	}

	// TODO enable this code block after implementing config macros (commands aliases)
	// for _, macro := range config.Macros {
	// 	if err := macro.CreateMounts(); err != nil {
	// 		return config, err
	// 	}
	// }

	return
}
