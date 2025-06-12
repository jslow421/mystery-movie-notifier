FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main .

FROM alpine:latest

# Install Chrome dependencies
RUN apk add --no-cache \
    chromium \
    ca-certificates \
    fontconfig \
    freetype \
    ttf-freefont &&
    rm -rf /var/cache/apk/*

# Create user for Chrome (Chrome won't run as root)
RUN addgroup -g 1000 chrome &&
    adduser -D -s /bin/sh -u 1000 -G chrome chrome

WORKDIR /app
COPY --from=builder /app/main .
USER chrome

ENV CHROME_BIN=/usr/bin/chromium-browser
ENV CHROME_PATH=/usr/bin/chromium-browser

CMD ["./main"]
