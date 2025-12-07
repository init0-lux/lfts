
FROM golang:1.22

WORKDIR /app

COPY . .

RUN go mod download
RUN go build -o lfts ./cmd/lfts

EXPOSE 9650

ENTRYPOINT ["./lfts", "start"]

