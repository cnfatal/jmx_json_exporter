#!/usr/bin/env bash

DOCKER_NAME=zookeeper_exporter
DOCKER_TAG=latest

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build
docker build -t ${DOCKER_NAME}:${DOCKER_TAG} .