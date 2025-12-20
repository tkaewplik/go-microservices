# Go Microservices Architecture

A comprehensive guide to understanding the microservices architecture, workflows, and features.

---

## 🏗️ System Overview

```mermaid
graph TB
    subgraph "Frontend"
        CS[🖥️ Client Service<br/>React App<br/>Port 3000]
    end

    subgraph "API Layer"
        GW[🚪 API Gateway<br/>Port 8080]
    end

    subgraph "Backend Services"
        AS[🔐 Auth Service<br/>Port 8081]
        PS[💳 Payment Service<br/>Port 8082]
        AN[📊 Analytics Service<br/>Port 8083]
    end

    subgraph "Databases"
        ADB[(Auth DB<br/>PostgreSQL)]
        PDB[(Payment DB<br/>PostgreSQL)]
    end

    subgraph "Message Brokers"
        KF[📨 Kafka KRaft<br/>Port 9092]
    end

    CS --> GW
    GW --> AS
    GW --> PS
    AS --> ADB
    PS --> PDB
    PS -->|Events| KF
    KF --> AN
```

---

## 🔄 User Authentication Flow

```mermaid
sequenceDiagram
    participant U as 👤 User
    participant C as 🖥️ Client
    participant G as 🚪 Gateway
    participant A as 🔐 Auth Service
    participant DB as 💾 Auth DB

    U->>C: Enter username/password
    C->>G: POST /auth/register or /login
    G->>A: Forward request
    A->>DB: Query/Insert user
    DB-->>A: User data
    A->>A: Hash password / Generate JWT
    A-->>G: Return { id, username, token }
    G-->>C: JWT Token
    C->>C: Store token locally
```

---

## 💰 Transaction Flow with Kafka Events

```mermaid
sequenceDiagram
    participant U as 👤 User
    participant C as 🖥️ Client
    participant G as 🚪 Gateway
    participant P as 💳 Payment Service
    participant DB as 💾 Payment DB
    participant K as 📨 Kafka
    participant AN as 📊 Analytics

    U->>C: Create transaction
    C->>G: POST /payment/transactions<br/>(with JWT)
    G->>P: Forward request
    P->>P: Validate JWT
    P->>DB: Check total < 1000
    P->>DB: Insert transaction
    DB-->>P: Transaction created
    P--)K: Publish "transaction.created"
    P-->>G: Return transaction
    G-->>C: Success response
    K--)AN: Consume event
    AN->>AN: Update statistics
```

---

## 📦 Service Details

### 🔐 Auth Service

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/register` | POST | Create new user account |
| `/login` | POST | Authenticate and get JWT |

**Features:**
- Password hashing with bcrypt
- JWT token generation (24h expiry)
- User data stored in PostgreSQL

---

### 💳 Payment Service

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/transactions` | POST | Create new transaction |
| `/transactions/list` | GET | Get user's transactions |
| `/transactions/pay` | POST | Mark all as paid |

**Features:**
- JWT authentication required
- Maximum 1000 total per user validation
- **Publishes events to Kafka** on create/pay

---

### 📊 Analytics Service

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | Health check |
| `/stats` | GET | Get aggregated statistics |

**Features:**
- Consumes events from Kafka
- Real-time statistics aggregation
- Tracks transactions by user

**Stats Response:**
```json
{
  "total_transactions": 10,
  "total_amount": 850.50,
  "total_paid_transactions": 5,
  "events_processed": 15,
  "unique_users": 3
}
```

---

## 🏛️ Code Architecture

Each Go service follows a **layered architecture** pattern:

```mermaid
graph TD
    subgraph "Service Structure"
        M[main.go<br/>Wiring & Server] --> H
        H[Handler Layer<br/>HTTP Endpoints] --> S
        S[Service Layer<br/>Business Logic] --> R
        R[Repository Layer<br/>Database Access] --> D[Domain Layer<br/>Models & Interfaces]
    end
```

**Directory Structure:**
```
service/
├── internal/
│   ├── domain/       # 📋 Interfaces & models
│   ├── repository/   # 💾 Database operations
│   ├── service/      # ⚙️ Business logic
│   ├── handler/      # 🌐 HTTP handlers
│   └── kafka/        # 📨 Event publishing
├── main.go           # 🚀 Entry point
└── Dockerfile
```

---

## 📨 Event-Driven Architecture

```mermaid
flowchart LR
    subgraph "Event Producer"
        PS[Payment Service]
    end

    subgraph "Message Broker"
        K[Kafka KRaft<br/>Topic: transactions]
    end

    subgraph "Event Consumers"
        AN[Analytics Service]
        FU[Future Services...]
    end

    PS -->|transaction.created| K
    PS -->|transaction.paid| K
    K --> AN
    K -.-> FU
```

**Event Types:**
| Event | Payload |
|-------|---------|
| `transaction.created` | `{transaction_id, user_id, amount, description, timestamp}` |
| `transaction.paid` | `{user_id, transactions_paid, timestamp}` |

---

## 🚀 Quick Start

```bash
# Start all services
docker compose up -d

# Access points
# Frontend:   http://localhost:3000
# API:        http://localhost:8080
# Analytics:  http://localhost:8083/stats
# RabbitMQ:   http://localhost:15672
```

---

## 🛠️ Tech Stack

| Component | Technology |
|-----------|------------|
| Backend | Go 1.25 |
| Frontend | React + TypeScript |
| Database | PostgreSQL 15 |
| Message Queue | Kafka 3.7 (KRaft) |
| Message Queue | RabbitMQ 3 |
| Gateway | Go HTTP Reverse Proxy |
| Auth | JWT (HS256) |
| Containerization | Docker + Docker Compose |
| CI/CD | GitHub Actions |
