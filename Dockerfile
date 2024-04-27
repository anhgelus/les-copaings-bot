FROM golang:1.22-alpine

WORKDIR /app

COPY . .

RUN apk add git

RUN go mod tidy && go build -o app .

ENV TOKEN=""

ENV FORCE_COMMAND_REGISTRATION="false"

CMD ./app -token $TOKEN -forge-command-registration $FORCE_COMMAND_REGISTRATION
