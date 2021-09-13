package queue

import (
	"encoding/json"
	"os/exec"
)

type JobStats struct {
	JobID     string
	ProcessID int64
	Status    JobStatus
}

type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusFailed    JobStatus = "failed"
	JobStatusSucceeded JobStatus = "succeeded"
)

type Job struct {
	ID    string   `json:"id"`
	Cmd   []string `json:"cmd"`
	Stats JobStats `json:"stats"`
}

func (j Job) String() string {
	b, _ := json.Marshal(j)

	return string(b)
}

func (j Job) Run() error {
	return exec.Command(j.Cmd[0], j.Cmd[1:]...).Run()
}
