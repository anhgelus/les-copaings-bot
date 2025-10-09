dev:
    podman network create db
    podman run -p 5432:5432 --rm --network db --name postgres --env-file .env -v ./data:/var/lib/postgresql/data -d postgres:alpine
    podman run -p 8080:8080 --rm --network db --name adminer -d adminer
    go run .

update:
    git stash
    git pull
    git stash pop
    go run .

stop:
    podman stop postgres adminer
    podman network rm db

clean-network:
    podman network rm db

build:
    go build -ldflags "-s" .
