FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /gtoc

FROM alpine:3.18

RUN apk add --no-cache git

WORKDIR /app
COPY --from=builder /gtoc /usr/local/bin/gtoc

ENTRYPOINT ["gtoc"]
