package handler

import (
	"assistant-plugin-go/hub"
	"log/slog"
	"strings"

	"github.com/spf13/viper"
)

type CommandPlugin interface {
	Command() []string
	OnCommand(keyword string, raw hub.Message) bool
}

func OnCommand(commands ...CommandPlugin) func(raw hub.Message) bool {
	prefix := viper.GetString("command.prefix")
	return func(raw hub.Message) bool {
		if !strings.HasPrefix(raw.Content, prefix) {
			return false
		}
		keyword1 := strings.SplitN(strings.TrimPrefix(raw.Content, prefix), " ", 2)[0]
		keyword2 := strings.SplitN(strings.TrimPrefix(raw.Content, prefix), "\n", 2)[0]
		for _, command := range commands {
			for _, key := range command.Command() {
				if key == keyword1 {
					slog.Info("OnCommand", "keyword", key)
					if command.OnCommand(keyword1, raw) {
						return true
					}
					break
				}
				if key == keyword2 {
					slog.Info("OnCommand", "keyword", key)
					if command.OnCommand(keyword2, raw) {
						return true
					}
					break
				}
			}
		}
		return false
	}
}
