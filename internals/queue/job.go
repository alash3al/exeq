package queue

import (
	"encoding/json"
	"os/exec"
)

type Job struct {
	ID  string   `json:"id"`
	Cmd []string `json:"cmd"`
}

func (j Job) String() string {
	b, _ := json.Marshal(j)

	return string(b)
}

func (j Job) Run() error {
	return exec.Command(j.Cmd[0], j.Cmd[1:]...).Run()
}
