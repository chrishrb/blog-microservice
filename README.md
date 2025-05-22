# 🚀 Blog Microservice Platform

A modern, scalable blog platform built with a microservices architecture using Go. **This is a learning project** designed to demonstrate microservices concepts and best practices.

## 📋 Overview

This project implements a complete blog platform with separate services for user management, post handling, and notifications. Each service is independently deployable and communicates through Kafka messaging to ensure reliability and scalability. The platform is built as an educational resource to explore and understand microservice architecture patterns.

## 🏗️ Architecture

The platform consists of the following microservices:

- **User Service**: Handles user registration, authentication, and profile management
- **Post Service**: Manages blog posts and comments
- **Notification Service**: Sends email notifications for account verification, password resets, etc.

### Communication

Services communicate asynchronously via Kafka for event-driven operations and directly via HTTP for synchronous API requests.

## ✨ Features

- **User Management**
  - Registration and account verification
  - Authentication with JWT
  - Password reset functionality
  - Role-based access control

- **Blog Content Management**
  - Create, read, update, and delete blog posts
  - Comment management

- **Notification System**
  - Email notifications
  - Customizable templates

## 💻 Tech Stack

- **Backend**: Go
- **Message Broker**: Kafka
- **API Documentation**: OpenAPI/Swagger
- **Containerization**: Docker
- **Observability**: OpenTelemetry, Jaeger

## 🚦 Getting Started

### Prerequisites

- Docker and Docker Compose
- Go 1.21 or higher

### Installation

1. Clone the repository:
```bash
git clone https://github.com/chrishrb/blog-microservice.git
cd blog-microservice
```

2. Start the services using Docker Compose:
```bash
docker-compose up -d
```

3. Access the services:
   - User Service API: http://localhost:9410
   - Post Service API: http://localhost:9411
   - Notification Service: http://localhost:9412
   - Swagger UI: http://localhost:8081
   - Kafka UI: http://localhost:8082
   - Mailpit (Email testing): http://localhost:8025
   - Jaeger UI (Tracing): http://localhost:16686

## 📝 API Documentation

The API is documented using OpenAPI/Swagger. You can access the documentation via Swagger UI at http://localhost:8081 after starting the services.

## 👨‍💻 Development

### Project Structure

```
blog-microservice/
├── user-service/         # User management service
├── post-service/         # Post and comment management service
├── notification-service/ # Notification delivery service
├── internal/             # Shared code between services
├── config/               # Configuration files
└── docker-compose.yaml   # Docker Compose configuration
```

### Building the Services

Build all services:

```bash
make build
```

## 🧪 Testing

Run the tests with:

```bash
make test
```

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

