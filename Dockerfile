# Use the official Go image as a base
FROM golang:1.24-alpine AS builder

# Set necessary environment variables
ENV CGO_ENABLED=0 GOOS=linux

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files to download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the Go application
RUN go build -o /kumparan-be-test ./cmd

# Use a minimal base image for the final stage
FROM alpine:latest

# Set the working directory
WORKDIR /root/

# Copy the built binary from the builder stage
COPY --from=builder /kumparan-be-test .
COPY --from=builder /app/migrations /app/migrations

# Expose the port the application will listen on
EXPOSE 9001

# Command to run the executable
CMD ["./kumparan-be-test"]
