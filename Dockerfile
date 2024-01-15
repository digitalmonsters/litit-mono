# Builder stage
FROM golang:1.21 AS builder

# Create and set the working directory
WORKDIR /app

# Copy go.mod and go.sum for dependency caching
COPY go.mod go.sum ./

# Set private repo access (if necessary)
ARG GITHUB_TOKEN
RUN if [ -n "$GITHUB_TOKEN" ]; then \
        git config --global url."https://${GITHUB_TOKEN}@github.com/".insteadOf "https://github.com/"; \
    fi

# Download dependencies
RUN go mod download

# Copy source files
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags='-w -s -extldflags "-static"' -a -o main cmd/main.go

# Final stage
FROM alpine:latest

# Add CA certificates
RUN apk --no-cache add ca-certificates

#Set the working directory
WORKDIR /app

#Copy the binary and configuration file from the builder stage
COPY --from=builder /app/main .
COPY --from=builder /app/config.json .

#Command to run the application
CMD ["/app/main"]
