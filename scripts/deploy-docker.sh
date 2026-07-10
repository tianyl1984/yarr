#!/bin/bash
cd "$(dirname "$0")/.."
docker compose -f deploy/docker-compose.yml up --build -d