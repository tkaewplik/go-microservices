# Go Microservices Tutorial

A simple tutorial for building a full-stack microservices application with Go backend services and React frontend.

## Tasks
- [ ] Write a scalable and maintainable codebase
- [ ] Implement proper error handling and logging
- [ ] Add unit tests for critical functionality
- [ ] Add rabbitMq and Kafka

## Architecture

This project demonstrates a microservices architecture with the following services:

- **API Gateway** (Port 8080) - Routes requests to backend services
- **Auth Service** (Port 8081) - Handles user authentication with JWT
- **Payment Service** (Port 8082) - Manages payment transactions
- **Client Service** (Port 3000) - React UI for interacting with the system

## Features

### Auth Service
- User registration and login
- JWT token generation
- Password hashing with bcrypt

### Payment Service
- Create transactions with user_id, amount, and description
- Automatic validation: maximum total amount of 1000 per user
- List all transactions for a user
- Pay all unpaid transactions for a user
- JWT authentication required for all endpoints

### API Gateway
- Routes requests to appropriate services
- CORS support for frontend integration
- Health check endpoint

### Client Service (React)
- User-friendly login/registration interface
- Transaction creation form with customizable auth header
- Real-time transaction list with status
- Visual indication of paid/unpaid transactions
- Total amount tracking with max limit display

## Prerequisites

- Go 1.21 or higher
- Node.js 18 or higher
- PostgreSQL 15 or higher
- Docker and Docker Compose (for containerized deployment)

## Quick Start with Docker Compose

1. Clone the repository:
```bash
git clone https://github.com/tkaewplik/go-microservices.git
cd go-microservices
```

2. Start all services:
```bash
docker-compose up -d
```

3. Migrate the database:
```bash
make migrate-up
```

4. Access the application:
- Frontend: http://localhost:3000
- API Gateway: http://localhost:8080
- Auth Service: http://localhost:8081
- Payment Service: http://localhost:8082

5. Stop all services:
```bash
docker-compose down
```

## Manual Setup (Development)

### 1. Setup Databases

Create two PostgreSQL databases:
```sql
CREATE DATABASE authdb;
CREATE DATABASE paymentdb;
```

### 2. Start Auth Service

```bash
cd auth-service
export DB_HOST=localhost
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=authdb
export JWT_SECRET=your-secret-key
export PORT=8081
go run main.go
```

### 3. Start Payment Service

```bash
cd payment-service
export DB_HOST=localhost
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=paymentdb
export JWT_SECRET=your-secret-key
export PORT=8082
go run main.go
```

### 4. Start API Gateway

```bash
cd gateway
export AUTH_SERVICE_URL=http://localhost:8081
export PAYMENT_SERVICE_URL=http://localhost:8082
export PORT=8080
go run main.go
```

### 5. Start Client Service

```bash
cd client-service
cp .env.example .env
npm install
npm start
```

## API Endpoints

### Auth Service (via Gateway: /auth/*)

#### Register
```bash
POST /auth/register
Content-Type: application/json

{
  "username": "testuser",
  "password": "password123"
}

Response:
{
  "id": 1,
  "username": "testuser",
  "token": "eyJhbGc..."
}
```

#### Login
```bash
POST /auth/login
Content-Type: application/json

{
  "username": "testuser",
  "password": "password123"
}

Response:
{
  "id": 1,
  "username": "testuser",
  "token": "eyJhbGc..."
}
```

### Payment Service (via Gateway: /payment/*)

All payment endpoints require JWT authentication via `Authorization: Bearer <token>` header.

#### Create Transaction
```bash
POST /payment/transactions
Authorization: Bearer <token>
Content-Type: application/json

{
  "user_id": 1,
  "amount": 100.50,
  "description": "Purchase of product X"
}

Response:
{
  "id": 1,
  "user_id": 1,
  "amount": 100.50,
  "description": "Purchase of product X",
  "is_paid": false,
  "created_at": "2024-01-15T10:30:00Z"
}
```

#### Get Transactions
```bash
GET /payment/transactions/list?user_id=1
Authorization: Bearer <token>

Response:
[
  {
    "id": 1,
    "user_id": 1,
    "amount": 100.50,
    "description": "Purchase of product X",
    "is_paid": false,
    "created_at": "2024-01-15T10:30:00Z"
  }
]
```

#### Pay All Transactions
```bash
POST /payment/transactions/pay?user_id=1
Authorization: Bearer <token>

Response:
{
  "message": "transactions paid successfully",
  "transactions_paid": 3
}
```

## Project Structure

```
.
├── auth-service/           # Authentication service
│   ├── main.go
│   └── Dockerfile
├── payment-service/        # Payment transaction service
│   ├── main.go
│   └── Dockerfile
├── gateway/                # API Gateway
│   ├── main.go
│   └── Dockerfile
├── client-service/         # React frontend
│   ├── src/
│   │   ├── components/
│   │   │   ├── LoginForm.tsx
│   │   │   ├── TransactionForm.tsx
│   │   │   └── TransactionList.tsx
│   │   ├── App.tsx
│   │   └── App.css
│   ├── Dockerfile
│   └── nginx.conf
├── pkg/                    # Shared packages
│   ├── database/           # Database utilities
│   ├── jwt/                # JWT utilities
│   └── middleware/         # HTTP middlewares
├── docker-compose.yml      # Docker Compose configuration
├── go.mod                  # Go module definition
└── README.md
```

## Environment Variables

### Auth Service
- `DB_HOST` - Database host (default: localhost)
- `DB_PORT` - Database port (default: 5432)
- `DB_USER` - Database user (default: postgres)
- `DB_PASSWORD` - Database password (default: postgres)
- `DB_NAME` - Database name (default: authdb)
- `JWT_SECRET` - Secret key for JWT signing (default: your-secret-key)
- `PORT` - Service port (default: 8081)

### Payment Service
- `DB_HOST` - Database host (default: localhost)
- `DB_PORT` - Database port (default: 5432)
- `DB_USER` - Database user (default: postgres)
- `DB_PASSWORD` - Database password (default: postgres)
- `DB_NAME` - Database name (default: paymentdb)
- `JWT_SECRET` - Secret key for JWT validation (default: your-secret-key)
- `PORT` - Service port (default: 8082)

### API Gateway
- `AUTH_SERVICE_URL` - Auth service URL (default: http://localhost:8081)
- `PAYMENT_SERVICE_URL` - Payment service URL (default: http://localhost:8082)
- `PORT` - Gateway port (default: 8080)

### Client Service
- `REACT_APP_API_URL` - API Gateway URL (default: http://localhost:8080)

## Testing the Application

1. **Register a new user** using the UI or API
2. **Login** to get an authentication token
3. **Create transactions** - try creating multiple transactions
4. **Test the 1000 limit** - create transactions totaling more than 1000 to see validation
5. **View transactions** - see all your transactions with their status
6. **Pay transactions** - use the "Pay All" button to mark all unpaid transactions as paid
7. **Test custom auth header** - modify the Authorization header in the transaction form

## Development Tips

- Use the browser console to see API requests and responses
- The React app includes a custom auth header field for testing different tokens
- Check service logs for debugging: `docker-compose logs -f <service-name>`
- Database data persists in Docker volumes

## Reference

This tutorial is inspired by [JordanMarcelino/learn-go-microservices](https://github.com/JordanMarcelino/learn-go-microservices) and demonstrates common microservices patterns in Go.

## License

MIT

