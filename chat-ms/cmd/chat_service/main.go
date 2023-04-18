package main

import (
	"database/sql"
	"davisbento/whats-gpt/chat-ms/configs"
	"davisbento/whats-gpt/chat-ms/internal/infra/repository"
	"davisbento/whats-gpt/chat-ms/internal/infra/web"
	"davisbento/whats-gpt/chat-ms/internal/infra/web/webserver"
	chatcompletion "davisbento/whats-gpt/chat-ms/internal/usecase/chat_completion"

	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/sashabaranov/go-openai"
)

func main() {
	configs, err := configs.LoadConfig(".")
	if err != nil {
		panic(err)
	}

	conn, err := sql.Open(configs.DBDriver, fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&multiStatements=true",
		configs.DBUser, configs.DBPassword, configs.DBHost, configs.DBPort, configs.DBName))
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	repo := repository.NewChatRepositoryMySQL(conn)
	client := openai.NewClient(configs.OpenAIApiKey)

	chatConfig := chatcompletion.ChatCompletionConfigInputDTO{
		Model:                configs.Model,
		ModelMaxTokens:       configs.ModelMaxTokens,
		Temperature:          float32(configs.Temperature),
		TopP:                 float32(configs.TopP),
		N:                    configs.N,
		Stop:                 configs.Stop,
		MaxTokens:            configs.MaxTokens,
		InitialSystemMessage: configs.InitialChatMessage,
	}

	useCase := chatcompletion.NewChatCompletionUseCase(repo, client)

	// streamChannel := make(chan chat_completion_stream.ChatCompletionOutputDTO)
	// useCaseStream := chat_completion_stream.NewChatCompletionStreamUseCase(repo, client, streamChannel)

	webServer := webserver.NewWebServer(":" + configs.WebServerPort)
	webServerChatHandler := web.NewWebChatGPTHandler(*useCase, chatConfig, configs.AuthToken)
	webServer.AddHandler("/chat", webServerChatHandler.Handle)

	fmt.Println("Server running on port " + configs.WebServerPort)
	webServer.Start()
}