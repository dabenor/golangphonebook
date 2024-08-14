# Dockerfile for standardization across platforms, from Docker documentation
# syntax=docker/dockerfile:1
FROM golang:1.23

# Set destination for COPY
WORKDIR /app

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY *.go ./

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /docker-phonebook

# Bind to a TCP Port
EXPOSE 8080

# Run
CMD ["/docker-phonebook"]