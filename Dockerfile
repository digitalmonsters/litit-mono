FROM golang:1.21 as builder

ARG GITHUB_TOKEN

RUN git config --global url."https://${GITHUB_TOKEN}@github.com/".insteadOf "https://github.com/"

# Create and set the working directory
WORKDIR /app

# Copy go.mod and go.sum for dependency caching
COPY go.mod go.sum ./

ENV GOPRIVATE=github.com/digitalmonsters*

RUN go mod download

COPY ./config.json ./config.json

# Copy source files
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build  -ldflags='-w -s -extldflags "-static"' -a -o main cmd/main.go

FROM alpine:latest

USER root

RUN apk --no-cache add ca-certificates

#Set the working directory
WORKDIR /app/

#Copy the binary and configuration file from the builder stage
COPY --from=builder /app/main /app/main
COPY --from=builder /app/config.json /app/config.json

RUN chmod +x /app/main

#Command to run the application
CMD ["/app/main"]
