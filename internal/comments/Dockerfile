############################
# STEP 1 build executable binary
############################
FROM golang:1.18-buster AS builder

# Create appuser.
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

# Configure work dir and copy files
WORKDIR /app
COPY . .

# Configure private module access
ADD priv/.netrc /root/.netrc
ENV GOPRIVATE=github.com/digitalmonsters/*

# Before building, tidy up the go.mod and go.sum files
RUN go mod tidy -e

COPY ./config.json /go/bin/

# Build the binary.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build  -ldflags='-w -s -extldflags "-static"' -a -o /go/bin/main cmd/main.go

# Run the binary.
CMD ["/go/bin/main"]
