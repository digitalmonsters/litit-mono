############################
# STEP 1: Build the executable binary
############################
FROM golang:1.18-buster AS builder

# Create a non-root user for running the application
ENV USER=appuser
ENV UID=10001

RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    "${USER}"

# Configure the working directory and copy files
WORKDIR /app
COPY . .

ADD priv/.netrc /root/.netrc
ENV GOPRIVATE=github.com/digitalmonsters/*


# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags='-w -s -extldflags "-static"' -a -o /app/main cmd/main.go

############################
# STEP 2: Create the final lightweight image
############################
FROM scratch

# Copy the binary from the builder stage
COPY --from=builder /app/main /app/main

# Copy the config file
COPY ./config.json /app/config.json

# Use the non-root user
USER ${USER}

# Set the working directory
WORKDIR /app

# Run the binary
CMD ["/app/main"]