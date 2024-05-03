package aic

import (
	"context"

	ollama "github.com/ollama/ollama/api"
)

type Ollama struct {
	*ollama.Client
	stream *bool
}

func NewOllamaClient() (*Ollama, error) {
	c, err := ollama.ClientFromEnvironment()
	if err != nil {
		return nil, err
	}
	f := false
	return &Ollama{
		Client: c,
		stream: &f,
	}, nil
}

func (o *Ollama) Converse(req *ollama.ChatRequest) *ollama.ChatResponse {
	ch := make(chan *ollama.ChatResponse)
	o.Chat(context.Background(), req, func(cr ollama.ChatResponse) error {
		ch <- &cr
		return nil
	})
	return <-ch
}

func (o *Ollama) HasModel(modelName string) (bool, error) {
	response, err := o.List(context.Background())
	if err != nil {
		return false, err
	}
	for _, model := range response.Models {
		if model.Model == modelName {
			return true, nil
		}
	}
	return false, nil
}
