Web Analyzer API (Go)

📦 Project Structure
.
├──-main.go        # Entry point
├── middleware/
│   └── middleware.go
├── models/
│   └── fetch.go
├── service/
│       └── analyzer.go 
│       └── fetch_service.go 
│       └── fetcher.go 
│       └── fetch_service_test.go 
├── go.mod
├── go.sum
├── Dockerfile
└── README.md

This project is a simple Go web service that fetches and analyzes web page content.

🚀 Run with Docker
1. Build Docker Image

docker build -t web-analyzer:latest .

2. Run Container
docker run -it --rm -p 8080:8080 web-analyzer:latest

Now the API is available at: http://localhost:8080/

<img width="1448" height="89" alt="image" src="https://github.com/user-attachments/assets/8bc5eafa-64ff-4330-b910-d3899b956a27" />

🔥 Example Request

Send a POST request with a JSON body containing the target URL:

curl --location 'localhost:8080/api/fetch' \
--header 'Content-Type: application/json' \
--data '{
    "url":"https%3A%2F%2Fwww.linkedin.com%2Ffeed%2F"
}'

Send a GET request with your sample url:

curl --location 'localhost:8080/api/fetch?url=https%3A%2F%2Fwww.linkedin.com%2Ffeed%2F'



🛠 Development (without Docker)

Go to the main directory via the command prompt and execute "go run main.go" or "go run ."
