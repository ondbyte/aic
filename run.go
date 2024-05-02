package aic

import (
	"context"
	"net/http"

	ollama "github.com/ollama/ollama/api"
	"github.com/ondbyte/ogo"
	cfg "github.com/spf13/viper"
)

func Run() {
	cfg.SetConfigName("cfg")
	cfg.SetConfigType("yaml")
	cfg.AddConfigPath(".")
	err := cfg.ReadInConfig()
	if err != nil {
		panic(err)
	}

	cfg.GetString("AI_MODEL_FILE")

	ollamac, err := NewOllamaClient()
	if err != nil {
		panic(err)
	}
	ollamac.Create(context.Background(),&ollama.CreateRequest{
		Model: cfg.GetString(),
	})
	messages := []ollama.Message{
		ollama.Message{
			Role:    "system",
			Content: "Provide very brief, concise responses",
		},
		ollama.Message{
			Role:    "user",
			Content: "Name some unusual animals",
		},
		ollama.Message{
			Role:    "assistant",
			Content: "Monotreme, platypus, echidna",
		},
		ollama.Message{
			Role:    "user",
			Content: "which of these is the most dangerous?",
		},
	}
	chatReq := &ollama.ChatRequest{
		Model:    "phi3",
		Messages: messages,
	}
	resp := ollamac.Converse(chatReq)
	
	ogoServer := ogo.New(nil)
	type Request struct {
		Msg string `json:"msg"`
	}
	type Response struct {
		Err string `json:"err"`
	}
	ogo.SetupHandler[Request, Response](
		ogoServer,
		"POST",
		"/chat",
		func(v *ogo.RequestValidator[Request, Response], reqData *Request) {
			v.Body(reqData, func(body *ogo.RequestBody) {
				body.MediaType(ogo.Json)
				body.Required(http.StatusTeapot, "invalid data")
			})
		},
		func(validatedStatus int, validatedErr string) (resp *ogo.Response[Response]) {
			if validatedErr != "" {
				return &ogo.Response[Response]{
					Status:    validatedStatus,
					MediaType: ogo.Json,
					Body: &Response{
						Err: validatedErr,
					},
				}
			}
			return nil
		},
		func(reqData *Request) (resp *ogo.Response[Response]) {

		},
	)
}
