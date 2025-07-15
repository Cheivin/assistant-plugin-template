package plugin

import (
	"assistant-plugin-go/hub"
	"log/slog"
)

type Echo struct {
}

func (p Echo) Command() []string {
	return []string{"echo"}
}

func (p Echo) OnCommand(keyword string, raw hub.Message) bool {
	_, err := hub.Client.SendText(raw.GID, raw.Content)
	if err != nil {
		slog.Error("Echo error", slog.Any("err", err))
	}
	return true
}
