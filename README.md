Web Analyzer API (Go)

Pre requirements
1. Install Docker Desktop
2. Install Golang 1.24 version
   Note: If you have any other golang latest version it should be updated the first line of the docker file. But not tested this scenario.

   FROM golang:1.24-alpine AS builder

📦 Project Structure

<img width="244" height="269" alt="image" src="https://github.com/user-attachments/assets/03ad2ec7-7fba-4216-a651-3dc415d11e06" />


This project is a simple Go web service that fetches and analyzes web page content.

🚀 Run with Docker
1. Build Docker Image

    docker build -t web-analyzer:latest .

2. Run Container
    docker run -it --rm -p 8080:8080 web-analyzer:latest

Now the API is available at: http://localhost:8080/

<img width="1448" height="89" alt="image" src="https://github.com/user-attachments/assets/8bc5eafa-64ff-4330-b910-d3899b956a27" />

🔥 Example Request


Send a GET request with your sample url:

    curl --location 'localhost:8080/api/fetch?url=https%3A%2F%2Fwww.linkedin.com%2Ffeed%2F'



🛠 Development (without Docker)


Go to the main directory via the command prompt and execute below commands
1. go mod tidy     -- install dependancies
2. go run main.go or go run .

 
🧪 Run Unit Tests

1. Open command prompt from the main folder
2. Go into the service folder (cd service)
3. Execute go test ./...

# Possible Improvements
 1. Analyzed responses can be saved in a database and then existing responses can be retrieved without doing analysis again.
 2. Authentication can be applied to avoid accessing unknown users.
 3. Parallel processing can be applied for treating multiple requests using WaitGroup.
