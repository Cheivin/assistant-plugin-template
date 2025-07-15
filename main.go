package main

import (
	"assistant-plugin-go/handler"
	"assistant-plugin-go/hub"
	"assistant-plugin-go/hub/sse"
	"assistant-plugin-go/plugin"
	"context"
	"errors"
	"github.com/spf13/viper"
	"log/slog"
	"os"
	"os/signal"
)

func init() {
	viper.SetDefault("ASSISTANT_HOST", "127.0.0.1")
	viper.SetDefault("ASSISTANT_PORT", 28003)
	viper.SetDefault("ASSISTANT_USERNAME", "")
	viper.SetDefault("ASSISTANT_USERNAME", "")
	// 用于识别命令的前缀
	viper.SetDefault("PREFIX_COMMAND", "#")

	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			slog.Error("Config file not found", "err", err)
		} else {
			panic(err)
		}
	}
	viper.AutomaticEnv()
}
func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	// 初始化客户端
	var options []sse.Option
	if viper.GetString("ASSISTANT_USERNAME") != "" && viper.GetString("ASSISTANT_PASSWORD") != "" {
		options = append(options, sse.WithAuth(viper.GetString("ASSISTANT_USERNAME"), viper.GetString("ASSISTANT_PASSWORD")))
	}
	hub.Init(sse.New(ctx, viper.GetString("ASSISTANT_HOST"), viper.GetInt("ASSISTANT_PORT"), options...))

	go func() {
		slog.Info("SSEClient connected")
		hub.Client.Listen(handler.Handle(
			handler.OnCommand(plugin.Echo{}),
		))
	}()
	<-ctx.Done()
}
