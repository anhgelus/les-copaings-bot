FROM golang:1.22-alpine

WORKDIR /app

COPY . .

RUN apk add git

RUN go mod tidy && go build -o app .

ENV TOKEN=""

CMD ./app -token $TOKEN
