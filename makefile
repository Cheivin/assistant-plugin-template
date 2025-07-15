.DEFAULT_GOAL := build
.PHONY : build

IMAGE_NAME:=assistant-plugin-go

build:
	go build -ldflags="-s -w" -o app ./
docker:
	docker build -f build/Dockerfile -t ${IMAGE_NAME} .
