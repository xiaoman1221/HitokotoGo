FROM golang:1.26 AS builder

ENV GOPROXY=https://goproxy.cn,direct

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o hitokotogo .

FROM alpine:3.20

WORKDIR /app

COPY --from=builder /app/hitokotogo /app/hitokotogo
COPY --from=builder /app/frontend /app/frontend
COPY --from=builder /app/.env.exp /app/.env

EXPOSE 8080

CMD ["./hitokotogo"]
