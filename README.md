# Chirpy

An HTTP server with RESTful API. 

## Features

- RESTful API
- Database storage (PostgreSQL)
- Authentication with JSON Web Tokens
- API tokens for authorization
- Refresh tokens
- Middleware
- Webhooks
- Unit testing
- Health check
- Metrics

## Further Improvements

- API documentation
- Dockerfile
- Integration testing
- API testing with scripts
- Continuous Integration with GitHub Actions
- Containerization with Docker & Docker Compose

## API Resources Overview

### Chirps

Get, create and delete chirps (short messages). 

**Endpoint**: `POST /api/chirps`

**Request**:
- **Headers**:
    - `Content-Type: application/json`
    - `Authorization: Bearer <JWT>`
- **Body**:
```json
{ "body": "message" }
```

**Response**:
- Status codes: 
    - 201 Created
    - 400 Bad Request
- **Body**:
```json
{ 
  "id": "uuid",
  "created_at": "timestamp",
  "updated_at": "timestamp",
  "body": "message",
  "user_id": "uuid",
}
```

- `GET /api/chirps`
- `GET /api/chirps/{chirpID}`
- `DELETE /api/chirps/{chirpID}`

### Users

- `POST /api/users`
- `PUT /api/users`
- `POST /api/login`
- `POST /api/refresh`
- `POST /api/revoke`

### Web Application

- `GET /app/`

### Utility

**Endpoint**: `POST /api/polka/webhooks`
**Request**:
- **Headers**:
    - `Content-Type: application/json`
    - `Authorization: ApiKey <token>`
- **Body**:
```json
{
    "event": "user.upgraded",
    "data": {
        "user_id": "uuid"
    }
}
```

- `POST /admin/reset`
- `POST /admin/metrics`
- `POST /api/healthz`
- `POST /api/validate_chirp`

## Prerequisites

- Running instance of PostgreSQL server
- Environmental variables:
```env
DB_URL="postgres://username:password@localhost:5432/chirpy"
PLATFORM="dev"
AUTH_SECRET="secret"
POLKA_KEY="api_secret"
```
