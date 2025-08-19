
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o web-analyzer .


FROM alpine:3.19


WORKDIR /app


COPY --from=builder /app/web-analyzer .

EXPOSE 8080


CMD ["./web-analyzer"]
