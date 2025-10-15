FROM golang:1.25.3 AS build
WORKDIR /app
ENV GOTOOLCHAIN=auto
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/litit ./cmd/api

FROM gcr.io/distroless/base-debian12
COPY --from=build /bin/litit /litit
EXPOSE 8080
ENTRYPOINT ["/litit"]