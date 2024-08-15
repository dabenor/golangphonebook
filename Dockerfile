# Dockerfile for standardization across platforms, from Docker documentation
# syntax=docker/dockerfile:1
FROM golang:1.23

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download Go modules
RUN go mod download

# Copy the entire project directory into the container
COPY . .

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux go build -o /docker-phonebook

# Expose the port the app will run on
EXPOSE 8080

# Run the application
CMD ["/docker-phonebook"]
