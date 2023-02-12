# Use an official Golang image as the base image
FROM golang:latest

# Set the working directory in the container to /app
WORKDIR /app

# Copy the current directory contents into the container at /app
COPY . .

# Build the Go app
RUN go build -o main .

# Expose port 8080 to the host
EXPOSE 8080

# Specify the command to run the app
CMD ["./main"]