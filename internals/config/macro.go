package config

import (
	"bytes"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/alash3al/exeq/pkg/utils"
)

type MacroConfig struct {
	Name    string `hcl:"name,label"`
	Command string `hcl:"command"`
	Mounts  []struct {
		Filename string `hcl:"filename,label"`
		Content  string `hcl:"content"`
	} `hcl:"mount,block"`
}

func (macro *MacroConfig) split() []string {
	return utils.SplitSpaceDelimitedString(strings.TrimSpace(macro.Command))
}

func (macro *MacroConfig) ParseAndSplit(ctx map[string]string) ([]string, error) {
	command, err := macro.parseCommand(ctx)
	if err != nil {
		return nil, err
	}

	return utils.SplitSpaceDelimitedString(strings.TrimSpace(command)), nil
}

func (macro *MacroConfig) validateCommand() error {
	cmdParts := macro.split()

	if len(cmdParts) < 1 {
		return fmt.Errorf("there is no command (or command is empty) for the macro (%s)", macro.Name)
	}

	if _, err := exec.LookPath(cmdParts[0]); err != nil {
		return err
	}

	return nil
}

func (macro *MacroConfig) setupMounts() (err error) {
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

func (macro *MacroConfig) parseCommand(ctx map[string]string) (string, error) {
	var out bytes.Buffer

	tpl, err := template.New(macro.Name).Parse(macro.Command)
	if err != nil {
		return "", err
	}

	if err := tpl.Execute(&out, struct {
		Args map[string]string
	}{ctx}); err != nil {
		return "", err
	}

	return out.String(), nil
}
