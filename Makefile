# SPDX-License-Identifier: GPL-3.0-or-later

IMAGE ?= fusion-content
TAG   ?= local

.PHONY: build test tidy docker-build run

build:
	go build ./...

test:
	go test ./...

tidy:
	go mod tidy

docker-build:
	eval $$(minikube docker-env) && docker build -t $(IMAGE):$(TAG) .

run:
	go run ./cmd/server/
