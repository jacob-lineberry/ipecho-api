# ipecho-api

[![Go Version](https://img.shields.io/github/go-mod/go-version/jacob-lineberry/ipecho-api)](https://github.com/jacob-lineberry/ipecho-api)
[![Build Status](https://github.com/jacob-lineberry/ipecho-api/actions/workflows/deploy.yaml/badge.svg)](https://github.com/jacob-lineberry/ipecho-api/actions/workflows/deploy.yaml)
[![License](https://img.shields.io/github/license/jacob-lineberry/ipecho-api)](https://github.com/jacob-lineberry/ipecho-api/blob/main/LICENSE)

Simple question, simple answer: what's my IP? That's what this service does.

**ipecho-api** is a lightweight, performant "What is my IP?" REST API service built with Go and deployed on Google Cloud Run. It's designed to be fast, secure, and easily consumable by humans and machines alike (especially via `curl`).

## Usage

You can use the service with any HTTP client. It's particularly useful with `curl` to quickly check your public IP from the terminal.

```bash
# Get your IP (auto-detect IPv4 or IPv6)
curl https://ipecho.dev

# Force IPv4
curl -4 https://ipecho.dev

# Force IPv6
curl -6 https://ipecho.dev

# Get IP in JSON format
curl https://ipecho.dev/json

# Force IPv4/IPv6 for JSON
curl -4 https://ipecho.dev/json
curl -6 https://ipecho.dev/json
```

## Features

- **Blazing Fast**: Written in Go 1.25.5 with minimal overhead and a ~12MB container footprint.
- **Dual Stack Support**: Fully compatible with both IPv4 and IPv6 flags in `curl`.
- **Production Ready**: Includes graceful shutdown, health checks, and request timeouts.
- **Rate Limited**: Built-in protection (120 req/min per IP) to ensure service availability.
- **Secure**: Runs as a non-root user in a minimal distroless container.
- **CI/CD Integrated**: Automated deployments to Google Cloud Run with provenance and SBOM.

## API Reference

### `GET /`
Returns the client's public IP address as plain text. This is the primary endpoint for terminal users.

- **Method**: `GET`
- **Response Type**: `text/plain; charset=utf-8`
- **Example Output**: `203.0.113.42`

### `GET /json`
Returns the client's public IP address wrapped in a JSON object.

- **Method**: `GET`
- **Response Type**: `application/json; charset=utf-8`
- **Example Output**: `{"ip": "203.0.113.42"}`

### `GET /health`
A simple health check endpoint used for monitoring and Cloud Run probes.

- **Method**: `GET`
- **Response Type**: `text/plain; charset=utf-8`
- **Output**: `ok`

## Local Development

### Prerequisites
- **Go**: Version 1.25.5 or higher.

### Running Locally
1. Clone the repository:
   ```bash
   git clone https://github.com/jacob-lineberry/ipecho-api.git
   cd ipecho-api
   ```
2. Start the server:
   ```bash
   go run ./cmd/server
   ```
   The server will start on `localhost:8080` by default.

### Testing
Once the server is running, you can test it from another terminal:
```bash
# Test plaintext endpoint
curl http://localhost:8080

# Test JSON endpoint
curl http://localhost:8080/json
```

## Docker

If you prefer containers, you can build and run the service locally using Docker.

### Build
```bash
docker build -t ipecho-api .
```

### Run
```bash
docker run -p 8080:8080 ipecho-api
```

The final image uses a multi-stage build and is based on `gcr.io/distroless/static-debian12`, resulting in a highly secure and minimal image (~12MB).

## Deployment (Cloud Run)

The project is configured for automated deployment via GitHub Actions.

- **Workflow**: `.github/workflows/deploy.yaml`
- **Target**: Google Cloud Run (`us-central1`)
- **Authentication**: Uses Workload Identity Federation (WIF).
- **Secrets Required**: 
  - `GCP_PROJECT_ID`
  - `WIF_PROVIDER`
  - `WIF_SERVICE_ACCOUNT`

## Architecture

The service is designed to work efficiently behind a load balancer.

1. **IP Extraction**: Custom middleware extracts the real client IP from the `X-Forwarded-For` header set by Cloud Run. It correctly identifies the original requester by taking the leftmost IP address.
2. **Context-Based**: The extracted IP is stored in the request context. This ensures that the rate limiter (using `go-chi/httprate`) and handlers are all looking at the same verified IP address.
3. **Graceful Shutdown**: The server listens for `SIGTERM` (sent by Cloud Run before scaling down) and `SIGINT`, allowing 10 seconds for in-flight requests to complete.
4. **Rate Limiting**: Rate limiting is applied per-instance based on the extracted client IP.

## Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | The port the server listens on | `8080` |

*Note: Rate limiting is configured for 120 requests per minute per IP address.*

## License

This project is licensed under the [MIT License](LICENSE).
