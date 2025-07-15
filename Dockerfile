# Stage 1: Build the application
FROM golang:1.23-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files to download dependencies
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy the rest of the application source code
COPY . .

# Build the application.
# CGO_ENABLED=0 is important for creating a static binary.
# -o /app/vybes-api specifies the output file name and location.
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/vybes-api ./cmd/api

# Stage 2: Create the final, minimal image
FROM gcr.io/distroless/static-debian11

# Set the working directory
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/vybes-api .

# Expose the port the app runs on
EXPOSE 8080

# Set the entrypoint for the container
ENTRYPOINT ["/app/vybes-api"]