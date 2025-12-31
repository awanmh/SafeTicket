# SafeTicket: High-Concurrency Ticket Booking System

**SafeTicket** is a high-performance backend simulation of a Ticket Flash Sale (War Tiket) written in Golang. It serves as a Proof of Concept (PoC) for handling extreme concurrency and preventing race conditions using PostgreSQL transactions and **PESSIMISTIC LOCKING** (`SELECT ... FOR UPDATE`).

## Tech Stack

- **Golang**: Gin Framework
- **Database**: PostgreSQL (pgx driver)
- **Testing**: K6 (Load Testing)
- **Infrastructure**: Docker, Docker Compose

## Features

- **Unsafe Booking Endpoint**: Deliberately vulnerable to race conditions. Includes artificial delay to demonstrate overselling when multiple users access stock simultaneously.
- **Safe Booking Endpoint**: Secure implementation using `SELECT ... FOR UPDATE` to ensure atomic stock checking and updates. Guarantees 0 overselling even under heavy load.

## How to Run

### 1. Start Infrastructure

Start PostgreSQL and the Application using Docker Compose:

```bash
docker-compose up -d --build
```

### 2. Run Load Tests (K6)

We use K6 to simulate a "Flash Sale" scenario where minimal stock is attacked by hundreds of concurrent users.

#### Scenario: The "War Tiket" Simulation

- **Stock**: 1 Ticket (or 100)
- **Users**: 100+ Concurrent Users (VUs)
- **Mechanism**: All users attempt to buy simultaneously.

#### A. Test Unsafe Endpoint (Race Condition Demo)

```bash
# Reset Database to 1 Ticket
docker exec safeticket-db psql -U user -d safeticket -c "DELETE FROM bookings; UPDATE events SET total_tickets = 1 WHERE id = 1;"

# Attack with 10 concurrent users
k6 run --vus 10 --iterations 10 -e MODE=unsafe k6-script.js
```

**Result**: You will likely see **Overselling**. The database inventory will show a negative value (e.g., -9), meaning 10 tickets were sold despite only 1 being available.

#### B. Test Safe Endpoint (The Solution)

```bash
# Reset Database to 1 Ticket
docker exec safeticket-db psql -U user -d safeticket -c "DELETE FROM bookings; UPDATE events SET total_tickets = 1 WHERE id = 1;"

# Attack with 50 concurrent users (Proof of Durability)
k6 run --vus 50 --iterations 50 -e MODE=safe k6-script.js
```

**Result**:

- **Successful Bookings**: 1
- **Failed Bookings (409 Conflict)**: 49
- **Database Stock**: 0 (No negatives!)

> **Proof that locking mechanism prevents overselling under high load.**
>
> ![Proof of Safe Mode](https://place-holder.png "Screenshot of K6 Result showing 1 success and rest failures") > _(Replace this placeholder with actual screenshot if available)_

## Manual Run (Without Docker)

1. Start DB: `docker-compose up -d db`
2. Run App: `go run cmd/server/main.go`
3. Run K6: `k6 run k6-script.js`
