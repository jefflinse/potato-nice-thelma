# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /potato-nice-thelma ./cmd/server

# Runtime stage
FROM gcr.io/distroless/static-debian12:nonroot

LABEL maintainer="Jeff Linse"
LABEL description="Potato Nice Thelma - meme generator service"

COPY --from=builder /potato-nice-thelma /potato-nice-thelma

EXPOSE 8080

ENTRYPOINT ["/potato-nice-thelma"]
