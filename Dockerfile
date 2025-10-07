# Build stage
FROM golang:1.24.7-trixie AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN apt-get update && apt-get upgrade -y && go mod download
COPY . .
RUN go build -o open-kube-event-exporter main.go

# Runtime stage
FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=builder /app/open-kube-event-exporter .
USER nonroot:nonroot
ENTRYPOINT ["/app/open-kube-event-exporter"]