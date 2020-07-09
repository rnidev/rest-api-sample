FROM golang:1.12-alpine

RUN apk update && apk upgrade && \
    apk add --no-cache bash git openssh && \
    apk add build-base

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

RUN go mod verify

COPY . .

CMD CGO_ENABLED=0 go test ./...

RUN go build -o main .

EXPOSE 8080

ENTRYPOINT ["./main"]
