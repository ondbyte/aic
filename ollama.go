package aic

import (
	"context"

	ollama "github.com/ollama/ollama/api"
)

type Olllama struct {
	*ollama.Client
}

func NewOllamaClient() (*Olllama, error) {
	c, err := ollama.ClientFromEnvironment()
	if err != nil {
		return nil, err
	}
	return &Olllama{
		Client: c,
	}, nil
}

func (o *Olllama) Converse(req *ollama.ChatRequest) *ollama.ChatResponse {
	ch := make(chan *ollama.ChatResponse)
	o.Chat(context.Background(), req, func(cr ollama.ChatResponse) error {
		ch <- &cr
		return nil
	})
	return <-ch
}
