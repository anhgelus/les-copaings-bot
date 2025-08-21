#!/usr/bin/bash

podman network create db
podman run -p 5432:5432 --rm --network db --name postgres --env-file .env -v ./data:/var/lib/postgresql/data -d postgres:alpine
podman run -p 8080:8080 --rm --network db --name adminer -d adminer
