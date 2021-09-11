package queue

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

type Job struct {
	ID  string            `json:"id"`
	Cmd []string          `json:"cmd"`
	Env map[string]string `json:"env"`
}

func (j Job) String() string {
	b, _ := json.Marshal(j)

	return string(b)
}

func (j Job) Run() error {
	fmt.Println("exec>", strings.Join(j.Cmd, " "))
	cmd := exec.Command(j.Cmd[0], j.Cmd[1:]...)
	return cmd.Run()
}
