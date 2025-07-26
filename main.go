package main

import (
	"assistant-plugin-go/handler"
	"assistant-plugin-go/hub"
	"assistant-plugin-go/hub/sse"
	"assistant-plugin-go/plugin"
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"strings"

	"github.com/spf13/viper"
)

func init() {
	viper.SetConfigType("yaml")

	// 先检查环境变量中是否指定了配置文件路径
	var err error
	if configPath := os.Getenv("CONFIG_FILE"); configPath != "" {
		viper.SetConfigFile(configPath)
		err = viper.ReadInConfig()
	} else {
		// 环境变量未指定，使用默认路径
		viper.SetConfigName("config")
		viper.AddConfigPath(".")
		err = viper.ReadInConfig()
	}

	if err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			slog.Error("Config file not found", "err", err)
		} else {
			panic(err)
		}
	}
	slog.Info("加载配置文件", slog.Any("config", viper.ConfigFileUsed()))

	viper.SetDefault("command.prefix", "#")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	// 初始化客户端
	var options []sse.Option
	if viper.GetString("assistant.username") != "" && viper.GetString("assistant.password") != "" {
		options = append(options, sse.WithAuth(viper.GetString("assistant.username"), viper.GetString("assistant.password")))
	}
	hub.Init(sse.New(ctx, viper.GetString("assistant.host"), viper.GetInt("assistant.port"), options...))

	go func() {
		slog.Info("SSEClient connected")
		hub.Client.Listen(handler.Handle(
			handler.OnCommand(plugin.Echo{}),
		))
	}()
	<-ctx.Done()
}
