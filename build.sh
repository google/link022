#!/bin/bash

env GOOS=linux GOARCH=arm CC=arm-linux-gnueabi-gcc go build -o binary/link022_agent agent/agent.go