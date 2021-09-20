package queue

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"time"

	"golang.org/x/net/context"
)

type JobStats struct {
	JobID     string
	ProcessID int64
}

type Job struct {
	ID          string   `json:"id"`
	Cmd         []string `json:"cmd"`
	MaxExecTime time.Duration
}

func (j Job) String() string {
	b, _ := json.Marshal(j)

	return string(b)
}

func (j Job) Run() error {
	var cancelFn context.CancelFunc
	ctx := context.Background()
	if j.MaxExecTime > 0 {
		ctx, cancelFn = context.WithTimeout(ctx, j.MaxExecTime)
		defer cancelFn()
	}

	fmt.Println("exec>", j.Cmd)

	return exec.CommandContext(ctx, j.Cmd[0], j.Cmd[1:]...).Run()
}
