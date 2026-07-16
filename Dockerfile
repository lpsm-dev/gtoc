FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /gtoc

FROM alpine:3.22

RUN addgroup -S gtoc && adduser -S gtoc -G gtoc

WORKDIR /work
COPY --from=builder /gtoc /usr/local/bin/gtoc

USER gtoc

ENTRYPOINT ["gtoc"]
