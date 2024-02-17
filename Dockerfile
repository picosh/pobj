FROM golang:1.22-alpine

WORKDIR /app

COPY go.* .

RUN go mod download

COPY . .

RUN go build -o objx cmd/authorized_keys/main.go

CMD [ "/app/objx" ]
