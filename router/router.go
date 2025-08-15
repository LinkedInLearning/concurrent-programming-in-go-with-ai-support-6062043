package router

import (
	"context"
	"fmt"
	"math/rand"
	"sync/atomic"
	"time"

	"github.com/charmbracelet/log"
	"github.com/openai/openai-go"
	"github.com/taigrr/animals"
	"github.com/taigrr/colorhash"
)

type Client interface {
	CreateChatCompletion(ctx context.Context, req openai.ChatCompletionNewParams) (*openai.ChatCompletion, error)
}

type namedClient struct {
	client Client
	name   string
	color  string
}

type Router struct {
	clients []namedClient
	counter uint64
	logger  *log.Logger
}

func NewRouter(clients []Client, logger *log.Logger) *Router {
	if logger == nil {
		logger = log.Default()
	}

	namedClients := make([]namedClient, len(clients))
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	animalNames := animals.Names()

	for i, client := range clients {
		animal := animalNames[rng.Intn(len(animalNames))]
		hash := colorhash.HashString(animal)
		color := colorhash.CreateColor(hash)
		ansiColor := fmt.Sprintf("38;2;%d;%d;%d", color.GetRed(), color.GetGreen(), color.GetBlue())

		namedClients[i] = namedClient{
			client: client,
			name:   animal,
			color:  ansiColor,
		}
	}

	return &Router{
		clients: namedClients,
		logger:  logger,
	}
}

func (r *Router) CreateChatCompletion(ctx context.Context, req openai.ChatCompletionNewParams) (*openai.ChatCompletion, error) {
	if len(r.clients) == 0 {
		return nil, fmt.Errorf("no clients available")
	}

	index := atomic.AddUint64(&r.counter, 1) - 1
	selectedClient := r.clients[index%uint64(len(r.clients))]

	r.logger.Info(fmt.Sprintf("\033[%sm[ %s ]\033[0m used to handle request", selectedClient.color, selectedClient.name))

	return selectedClient.client.CreateChatCompletion(ctx, req)
}
