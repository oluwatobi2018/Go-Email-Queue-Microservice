# Email Queue Microservice

A production-ready Go microservice that accepts email jobs over HTTP, queues them, and processes them asynchronously using a worker system.

## Features

### Core Features
- **HTTP API** with POST /send-email endpoint
- **In-memory job queue** with configurable size
- **Multiple concurrent workers** (configurable)
- **Graceful shutdown** handling
- **Input validation** with proper error responses

### Bonus Features
- **Retry logic** with exponential backoff (up to 3 retries)
- **Dead letter queue** for permanently failed jobs
- **Prometheus metrics** for monitoring
- **Configurable workers** and queue size via environment variables
- **Health check** endpoint
- **Structured logging** with JSON format

## API Endpoints

### POST /api/v1/send-email
Submit an email job for processing.

**Request:**
```json
{
  "to": "user@example.com",
  "subject": "Welcome!",
  "body": "Thanks for signing up."
}
```

**Response:**
- `202 Accepted` - Job queued successfully
- `422 Unprocessable Entity` - Invalid input
- `503 Service Unavailable` - Queue is full or service is shutting down

### GET /api/v1/stats
Get queue statistics.

**Response:**
```json
{
  "queue_length": 5,
  "retry_queue_length": 2,
  "dead_letter_count": 1,
  "is_closed": false
}
```

### GET /api/v1/dead-letter
Get failed jobs in the dead letter queue.

### GET /health
Health check endpoint.

### GET /metrics
Prometheus metrics endpoint (port 9090).

## Configuration

Environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | 8080 | HTTP server port |
| `WORKER_COUNT` | 3 | Number of worker goroutines |
| `QUEUE_SIZE` | 100 | Maximum queue size |
| `MAX_RETRIES` | 3 | Maximum retry attempts |
| `METRICS_PORT` | 9090 | Metrics server port |

## Running the Service

### Using Go directly

```bash
# Install dependencies
go mod tidy

# Run the service
go run main.go

# Or build and run
go build -o email-service
./email-service
```

### Using Docker

```bash
# Build the image
docker build -t email-queue-service .

# Run the container
docker run -p 8080:8080 -p 9090:9090 email-queue-service

# With custom configuration
docker run -p 8080:8080 -p 9090:9090 \
  -e WORKER_COUNT=5 \
  -e QUEUE_SIZE=200 \
  email-queue-service
```

## Testing

### Send an email
```bash
curl -X POST http://localhost:8080/api/v1/send-email \
  -H "Content-Type: application/json" \
  -d '{
    "to": "test@example.com",
    "subject": "Test Email",
    "body": "This is a test email."
  }'
```

### Check statistics
```bash
curl http://localhost:8080/api/v1/stats
```

### Check dead letter queue
```bash
curl http://localhost:8080/api/v1/dead-letter
```

### Health check
```bash
curl http://localhost:8080/health
```

### Metrics
```bash
curl http://localhost:9090/metrics
```

## Architecture

The service is built with proper separation of concerns:

- **config/**: Configuration management
- **handlers/**: HTTP request handlers
- **models/**: Data structures and validation
- **queue/**: Job queue implementation with retry logic
- **service/**: Business logic for email processing
- **worker/**: Worker pool for concurrent job processing

## Monitoring

The service exposes Prometheus metrics:

- `email_queue_length`: Current number of jobs in queue
- `email_jobs_processed_total`: Total jobs processed (by status)

## Graceful Shutdown

The service handles SIGINT and SIGTERM signals:

1. Stops accepting new HTTP requests
2. Closes the job queue
3. Waits for all active workers to finish processing
4. Shuts down HTTP servers
5. Exits cleanly

## Production Considerations

- **Logging**: Structured JSON logging with configurable levels
- **Metrics**: Prometheus metrics for monitoring
- **Health checks**: Endpoint for load balancer health checks
- **Timeouts**: Configurable timeouts for HTTP operations
- **Error handling**: Comprehensive error handling with proper HTTP status codes
- **Graceful shutdown**: Clean shutdown process
- **Resource limits**: Configurable queue size and worker count
- **Docker ready**: Multi-stage Dockerfile for optimized container