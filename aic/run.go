package aic

import (
	"context"
	"net/http"
	"os"
	"time"

	ollama "github.com/ollama/ollama/api"
	"github.com/ondbyte/ogo"
	cfg "github.com/spf13/viper"
)

func Run() {
	defer cfg.SafeWriteConfig()
	cfg.SetConfigName("cfg")
	cfg.SetConfigType("yaml")
	cfg.AddConfigPath(".")
	cfg.SetDefault("ollama_model", "yphi3")
	cfg.SetDefault("ollama_host", "http://127.0.0.1:11434")
	err := cfg.ReadInConfig()
	if err != nil {
		panic(err)
	}
	//
	modelName := cfg.GetString("ollama_model")
	//
	err = cfg.ReadInConfig()
	if err != nil {
		panic(err)
	}

	//
	ollamaHost := cfg.GetString("ollama_host")
	os.Setenv("OLLAMA_HOST", ollamaHost)
	ollamac, err := ollama.ClientFromEnvironment()
	if err != nil {
		panic(err)
	}

	//check whether ollama is running
	_, err = ollamac.List(context.Background())
	if err != nil {
		panic(err)
	}
	streamEnabled := true

	ogoServer := ogo.NewServer(nil)
	type Request struct {
		Msg string `json:"msg,omitempty"`
	}
	type Response struct {
		Err string `json:"err,omitempty"`
		Msg string `json:"msg,omitempty"`
	}
	ogo.SetupHandler[Request, Response](
		ogoServer,
		"POST",
		"/chat",
		func(v *ogo.RequestValidator[Request, Response], reqData *Request) {
			v.Body(reqData, func(body *ogo.RequestBody) {
				body.MediaType(ogo.Json)
				body.Required(http.StatusTeapot, "body is missing")
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
			resultCh := make(chan *ollama.ChatResponse)
			timeOut, cancel := context.WithTimeout(context.TODO(), time.Second*16)
			err = ollamac.Chat(
				timeOut,
				&ollama.ChatRequest{
					Stream: &streamEnabled,
					Model:  modelName,
					Messages: []ollama.Message{
						ollama.Message{
							Role:    "user",
							Content: reqData.Msg,
						},
					},
				},
				func(cr ollama.ChatResponse) error {
					resultCh <- &cr
					cancel()
					return nil
				})
			if err != nil {
				return &ogo.Response[Response]{
					Status: http.StatusInternalServerError,
					Body: &Response{
						Err: err.Error(),
					},
				}
			}
			select {
			case result := <-resultCh:
				{
					return &ogo.Response[Response]{
						Status:    http.StatusOK,
						MediaType: ogo.Json,
						Body: &Response{
							Msg: result.Message.Content,
						},
					}
				}
			case <-timeOut.Done():
				{
					break
				}
			}
			return &ogo.Response[Response]{
				Status:    http.StatusInternalServerError,
				MediaType: ogo.Json,
				Body: &Response{
					Err: "timeout",
				},
			}
		},
	)
	ogoServer.Run(8765, func(info *ogo.ServerInfo) {
		info.Description("main server")
	})
}
