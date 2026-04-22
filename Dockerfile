FROM golang:1.26.0 AS builder

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o hitokotogo .

FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/hitokotogo /app/hitokotogo
COPY --from=builder /app/data /app/data
COPY --from=builder /app/.env.exp /app/.env

EXPOSE 8080

CMD ["./hitokotogo"]