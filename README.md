# Aether Node

Air Quality Management & Monitoring IoT Dashboard Backend.

## Stack

- **Go 1.22** + Echo v4
- **PostgreSQL 16** — master data (users, devices, locations, etc.)
- **InfluxDB 2.7** — telemetry time-series data
- **n8n** — workflow automation
- **Docker Compose** — full stack deployment

## Project Structure (SoA + Domain Layer)

```
aether-node/
├── cmd/server/main.go           # Entry point + DI wiring
├── internal/
│   ├── domain/                   # Interfaces (contracts)
│   │   ├── user/
│   │   ├── device/
│   │   ├── location/
│   │   ├── installation_point/
│   │   ├── apikey/
│   │   ├── auth/
│   │   └── telemetry/
│   ├── repository/               # PostgreSQL / InfluxDB implementations
│   │   ├── user/
│   │   ├── device/
│   │   └── ...
│   ├── service/                  # Business logic implementations
│   │   ├── user/
│   │   ├── device/
│   │   └── ...
│   └── handler/                  # HTTP handlers
├── pkg/response/                 # Standardized API response helpers
├── migrations/                   # SQL migration files
└── docker-compose.yml            # Full stack deployment
```

## Quick Start (Docker)

```bash
# 1. Copy environment file
cp .env.example .env

# 2. Start all services
docker-compose up -d

# 3. Run migrations (automatic on first start via init script)
# Or manually:
make migrate-up

# 4. Check health
curl http://localhost:8080/health
```

## Services

| Service | URL | Credentials |
|---------|-----|-------------|
| API | http://localhost:8080 | — |
| n8n | http://localhost:5678 | admin / admin123 |
| PostgreSQL | localhost:5432 | postgres / postgres123 |
| InfluxDB | http://localhost:8086 | admin / admin123 |

## Manual Development

```bash
# Install dependencies
go mod tidy

# Run migrations
make migrate-up

# Start server
go run ./cmd/server

# Build
go build -o aether-node ./cmd/server
```

## API Endpoints

### Auth (Public)
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/auth/register` | Register new user |
| POST | `/auth/login` | Login |
| POST | `/auth/forgot-password` | Forgot password |
| POST | `/auth/token/refresh` | Refresh token |
| POST | `/auth/logout` | Logout (JWT required) |

### Protected Routes (JWT Required)
| Resource | Endpoints |
|----------|-----------|
| User | `/user`, `/user/:guid`, `/user/list` |
| Device | `/device`, `/device/:guid`, `/device/list` |
| Location | `/location`, `/location/:guid`, `/location/list` |
| Installation Point | `/installation-point`, `/installation-point/:guid`, `/installation-point/:guid/relations`, `/installation-point/list` |
| API Key | `/apikey`, `/apikey/:guid`, `/apikey/list` |
| Telemetry Stream | `/stream`, `/stream/:device-sn` |
| Telemetry History | `/history/telemetry/:device-sn` |

### Telemetry Ingestion (API Key Auth)
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/telemetry` | Write telemetry data |

## Database Schema

All master data tables use **soft delete** pattern.

```
users              → auth, profile
devices            → device registry
locations          → location master
installation_points → device-location mapping
apikeys            → device API keys
refresh_tokens     → JWT refresh tokens
```

InfluxDB (separate):
```
telemetry          → time-series IoT readings (temperature, humidity, AQI, PM2.5, etc.)
```

## License

MIT
# test
