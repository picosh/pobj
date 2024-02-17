FROM golang:1.22-alpine

WORKDIR /app

COPY go.* .

RUN go mod download

COPY . .

RUN go build -o pobj cmd/authorized_keys/main.go

CMD [ "/app/pobj" ]
