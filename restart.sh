#!/usr/bin/bash

podman compose down && podman compose build && podman compose up -d