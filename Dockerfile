# syntax=docker/dockerfile:1

# Build
FROM golang:1.23 AS build-stage

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /imagine

# Final
FROM gcr.io/distroless/base-debian11

WORKDIR /
COPY --from=build-stage /imagine /imagine
USER nonroot:nonroot
ENTRYPOINT ["/imagine"]