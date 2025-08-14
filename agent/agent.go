package agent

import (
	"context"
)

// Agent defines the interface for all agents in the system.
// Agents process input through channels and produce output asynchronously.
type Agent interface {
	Start(ctx context.Context, input <-chan string, output chan<- string) error
}

// Config holds the configuration parameters for an agent.
type Config struct {
	Name   string
	Model  string
	Prompt string
}