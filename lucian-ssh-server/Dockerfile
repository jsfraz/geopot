# Use the official Golang image as the base image
FROM golang:1.22.0-alpine AS build

# Set the working directory inside the container
WORKDIR /app

# Copy the Go mod and sum files to the working directory
COPY go.mod go.sum ./

# Download and install dependencies
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the application
RUN go build -o lucian

# Start a new stage using a minimal Alpine image
FROM alpine:latest

# Set the working directory inside the container
WORKDIR /app

# Copy the built binary from the previous stage
COPY --from=build /app/lucian .

# Expose the port that the application will run on
EXPOSE 2222

# Command to run the application
CMD ["./lucian"]