# Start from the official Go image
FROM golang:1.17-alpine

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY ./cmd .

# Copy the default config file
COPY monitors.yaml /app/monitors.yaml

# Build the Go app
RUN go build -o health-checks .

# Command to run the executable
CMD ["./health-checks"]