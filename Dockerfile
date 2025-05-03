# Dockerfile
FROM golang:1.23 as builder

WORKDIR /app

ENV GOTOOLCHAIN=local

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /go-server ./cmd/server

# Final image
FROM gcr.io/distroless/static

COPY --from=builder /go-server /go-server

ENTRYPOINT ["/go-server"]
