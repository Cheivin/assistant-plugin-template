package handler

import "assistant-plugin-go/hub"

func Handle(fn ...func(raw hub.Message) bool) func(hub.Message) {
	return func(raw hub.Message) {
		for _, f := range fn {
			if f(raw) {
				return
			}
		}
	}
}
