package queue

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/kataras/golog"
	"golang.org/x/net/context"
)

type JobStats struct {
	JobID     string
	ProcessID int64
}

type Job struct {
	// User space input
	ID          string   `json:"id"`
	Cmd         []string `json:"cmd"`
	MaxExecTime time.Duration

	// System level input
	EnqueuedAt    time.Time
	StartedAt     time.Time
	FinishedAt    time.Time
	RetryAttempts int64
	Error         string
}

func (j Job) String() string {
	b, _ := json.Marshal(j)

	return string(b)
}

func (j Job) Run() error {
	if len(j.Cmd) < 1 {
		return fmt.Errorf("empty command specified")
	}

	var cancelFn context.CancelFunc

	ctx := context.Background()

	if j.MaxExecTime > 0 {
		ctx, cancelFn = context.WithTimeout(ctx, j.MaxExecTime)
		defer cancelFn()
	}

	golog.Info("running> ", strings.Join(j.Cmd, " "))

	var stderr bytes.Buffer

	cmd := exec.CommandContext(ctx, j.Cmd[0], j.Cmd[1:]...)
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error: %s, details: %s", err.Error(), stderr.String())
	}

	if stderr.Len() > 0 {
		return fmt.Errorf("error: %s", stderr.String())
	}

	return nil
}

func (j Job) IsRunning() bool {
	return j.FinishedAt.IsZero() && !j.StartedAt.IsZero()
}

func (j Job) HasFinished() bool {
	return !j.FinishedAt.IsZero()
}
