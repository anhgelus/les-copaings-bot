FROM docker.io/golang:1.24-alpine

WORKDIR /app

RUN apk add git

COPY . .

RUN go mod tidy && go build -o app .

ENV TOKEN=""
ENV TZ="Europe/Paris"

CMD ./app -token $TOKEN
