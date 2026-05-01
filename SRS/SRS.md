# SRS - Aether Node

## Software Requirements Specification
**Versi:** 1.0  
**Tanggal:** 29 April 2026  
**Produk:** Aether Node - Air Quality Management & Monitoring IoT Device  
**Base URL:** https://aether-be.wit-sby.id

---

## 1. Introduction

### 1.1 Purpose
Dokumen ini menjabarkan spesifikasi teknis lengkap untuk backend system **Aether Node** yang menangani:
- Authentication & Authorization
- Master Data Management (User, Role, Device, Location, Installation Point, API Key)
- Realtime Data Streaming via SSE
- Historical Data Query dari InfluxDB

### 1.2 Scope
Backend API service menggunakan Go + Echo Framework dengan:
- PostgreSQL untuk transactional data (CRUD operations)
- InfluxDB untuk time-series telemetry data
- SoA (Service-Oriented Architecture) pattern
- JWT + API Key untuk authentication

### 1.3 Definitions & Acronyms

| Term | Definition |
|------|------------|
| **GUID** | Global Unique Identifier - UUID v4 |
| **SSE** | Server-Sent Events - unidirectional real-time stream |
| **SoA** | Service-Oriented Architecture |
| **Soft Delete** | Data tidak dihapus, hanya set `deleted_at` timestamp |
| **JWT** | JSON Web Token untuk authentication |
| **API Key** | Key untuk device-to-server authentication |

---

## 2. Technical Architecture

### 2.1 Technology Stack

| Component | Technology |
|-----------|------------|
| Language | Go 1.22+ |
| Framework | Echo Framework v4 |
| SQL Generator | SQLC |
| Architecture | SoA (Service-Oriented Architecture) |
| Primary Database | PostgreSQL 16 |
| Time-Series DB | InfluxDB |
| Authentication | JWT (Access + Refresh Token) + API Key |
| Password Hashing | bcrypt |

### 2.2 Project Structure (SoA Pattern with Domain Layer)

```
aether-node/
├── cmd/
│   └── server/
│       └── main.go                    # Entry point, DI setup, route registration
│
├── internal/
│   ├── config/                        # Configuration management
│   │   └── config.go
│   │
│   ├── domain/                        # Domain entities & interfaces
│   │   ├── user/
│   │   │   ├── user_entity.go         # User domain entity
│   │   │   ├── user_repository.go     # User repository interface
│   │   │   └── user_service.go        # User service interface
│   │   ├── role/
│   │   │   ├── role_entity.go         # Role domain entity
│   │   │   ├── role_repository.go     # Role repository interface
│   │   │   └── role_service.go        # Role service interface
│   │   ├── device/
│   │   │   ├── device_entity.go       # Device domain entity
│   │   │   ├── device_repository.go   # Device repository interface
│   │   │   └── device_service.go      # Device service interface
│   │   ├── location/
│   │   │   ├── location_entity.go     # Location domain entity
│   │   │   ├── location_repository.go # Location repository interface
│   │   │   └── location_service.go    # Location service interface
│   │   ├── installation_point/
│   │   │   ├── installation_point_entity.go  # InstallationPoint entity
│   │   │   ├── installation_point_repository.go
│   │   │   └── installation_point_service.go
│   │   ├── apikey/
│   │   │   ├── apikey_entity.go       # API Key domain entity
│   │   │   ├── apikey_repository.go   # API Key repository interface
│   │   │   └── apikey_service.go      # API Key service interface
│   │   ├── auth/
│   │   │   ├── auth_entity.go         # Auth domain entities (tokens)
│   │   │   ├── auth_repository.go     # Auth repository interface
│   │   │   └── auth_service.go        # Auth service interface
│   │   └── telemetry/
│   │       ├── telemetry_entity.go    # Telemetry domain entity
│   │       ├── telemetry_repository.go # Telemetry repository interface
│   │       └── telemetry_service.go   # Telemetry service interface
│   │
│   ├── repository/                    # Repository implementations (impl)
│   │   ├── user/
│   │   │   └── user_repository_impl.go
│   │   ├── role/
│   │   │   └── role_repository_impl.go
│   │   ├── device/
│   │   │   └── device_repository_impl.go
│   │   ├── location/
│   │   │   └── location_repository_impl.go
│   │   ├── installation_point/
│   │   │   └── installation_point_repository_impl.go
│   │   ├── apikey/
│   │   │   └── apikey_repository_impl.go
│   │   ├── auth/
│   │   │   └── auth_repository_impl.go
│   │   └── telemetry/
│   │       └── telemetry_repository_impl.go
│   │
│   ├── service/                       # Service implementations (impl)
│   │   ├── user/
│   │   │   └── user_service_impl.go
│   │   ├── role/
│   │   │   └── role_service_impl.go
│   │   ├── device/
│   │   │   └── device_service_impl.go
│   │   ├── location/
│   │   │   └── location_service_impl.go
│   │   ├── installation_point/
│   │   │   └── installation_point_service_impl.go
│   │   ├── apikey/
│   │   │   └── apikey_service_impl.go
│   │   ├── auth/
│   │   │   └── auth_service_impl.go
│   │   └── telemetry/
│   │       └── telemetry_service_impl.go
│   │
│   ├── handler/                       # HTTP handlers (controllers)
│   │   ├── auth_handler.go
│   │   ├── user_handler.go
│   │   ├── role_handler.go
│   │   ├── device_handler.go
│   │   ├── location_handler.go
│   │   ├── installation_point_handler.go
│   │   ├── apikey_handler.go
│   │   └── telemetry_handler.go
│   │
│   └── dto/                          # Data Transfer Objects
│       ├── request/                   # Request DTOs
│       │   ├── auth/
│       │   │   ├── login_request.go
│       │   │   ├── register_request.go
│       │   │   └── refresh_token_request.go
│       │   ├── user_request.go
│       │   ├── device_request.go
│       │   ├── location_request.go
│       │   ├── installation_point_request.go
│       │   └── apikey_request.go
│       └── response/                  # Response DTOs
│           ├── auth_response.go
│           ├── user_response.go
│           ├── device_response.go
│           └── common_response.go
│
├── pkg/
│   ├── response/                      # Standardized API responses
│   │   └── response.go
│   ├── middleware/                    # Custom middleware
│   │   ├── auth.go
│   │   └── apikey.go
│   └── database/                     # Database connections
│       ├── postgres.go
│       └── influxdb.go
│
├── migrations/                        # SQL migrations
│   ├── 001_create_users.sql
│   ├── 002_create_roles.sql
│   ├── 003_create_devices.sql
│   ├── 004_create_locations.sql
│   ├── 005_create_installation_points.sql
│   └── 006_create_apikeys.sql
│
├── sqlc.yaml                          # SQLC configuration
├── go.mod
├── go.sum
└── README.md
```

### 2.2.1 Layer Structure Diagram

```
┌─────────────────────────────────────────────────────────┐
│                     Handler Layer                       │
│  (auth_handler, user_handler, device_handler, etc.)     │
└─────────────────────────┬───────────────────────────────┘
                          │ depends on interface
┌─────────────────────────▼───────────────────────────────┐
│                     Service Layer                        │
│     ┌─────────────────────┐    ┌─────────────────────┐   │
│     │  Service Interface  │    │ Service Implementation│  │
│     │  (user_service.go) │    │ (user_service_impl.go)│  │
│     └─────────────────────┘    └─────────────────────┘   │
└─────────────────────────┬───────────────────────────────┘
                          │ depends on interface
┌─────────────────────────▼───────────────────────────────┐
│                   Repository Layer                      │
│     ┌─────────────────────┐    ┌─────────────────────┐   │
│     │ Repository Interface│    │Repository Implementation│ │
│     │(user_repository.go)│    │(user_repository_impl.go)││
│     └─────────────────────┘    └─────────────────────┘   │
└─────────────────────────┬───────────────────────────────┘
                          │
┌─────────────────────────▼───────────────────────────────┐
│                    Domain Layer                        │
│              (Entities, Value Objects)                 │
│     ┌─────────────────────────────────────────────┐    │
│     │  user_entity.go  │ device_entity.go  │ etc │    │
│     └─────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────┘
```

### 2.2.2 Domain Entity Pattern

**Example: Device Module**

```go
// domain/device/device_entity.go
package device

import "time"

type Device struct {
    GUID         string     `json:"guid"`
    Type         string     `json:"type"`
    SerialNumber string     `json:"serial_number"`
    Alias        string     `json:"alias"`
    Notes        string     `json:"notes"`
    IsActive     bool       `json:"is_active"`
    CreatedAt    time.Time  `json:"created_at"`
    UpdatedAt    time.Time  `json:"updated_at"`
    DeletedAt    *time.Time `json:"deleted_at,omitempty"`
}

// domain/device/device_repository.go
package device

type DeviceRepository interface {
    Create(ctx context.Context, device *Device) error
    GetByGUID(ctx context.Context, guid string) (*Device, error)
    List(ctx context.Context, params ListParams) ([]*Device, int64, error)
    Update(ctx context.Context, device *Device) error
    Delete(ctx context.Context, guid string) error
    ExistsBySerialNumber(ctx context.Context, serialNumber string) (bool, error)
}

// domain/device/device_service.go
package device

type DeviceService interface {
    CreateDevice(ctx context.Context, req *CreateDeviceRequest) (*Device, error)
    GetDevice(ctx context.Context, guid string) (*Device, error)
    ListDevices(ctx context.Context, params *ListRequest) (*ListResponse, error)
    UpdateDevice(ctx context.Context, guid string, req *UpdateDeviceRequest) (*Device, error)
    DeleteDevice(ctx context.Context, guid string) error
}
```

### 2.3 Database Schema

#### PostgreSQL Tables

**Table: users**
```sql
CREATE TABLE users (
    guid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    role_guid UUID REFERENCES roles(guid),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);
```

**Table: roles**
```sql
CREATE TABLE roles (
    guid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) UNIQUE NOT NULL,
    permissions JSONB DEFAULT '[]',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);
```

**Table: devices**
```sql
CREATE TABLE devices (
    guid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type VARCHAR(100) NOT NULL,
    serial_number VARCHAR(100) UNIQUE NOT NULL,
    alias VARCHAR(255),
    notes TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);
```

**Table: locations**
```sql
CREATE TABLE locations (
    guid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);
```

**Table: installation_points**
```sql
CREATE TABLE installation_points (
    guid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    device_guid UUID REFERENCES devices(guid),
    location_guid UUID REFERENCES locations(guid),
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);
```

**Table: api_keys**
```sql
CREATE TABLE api_keys (
    guid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key_hash VARCHAR(255) UNIQUE NOT NULL,
    notes TEXT,
    expire_date TIMESTAMP NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);
```

**Table: refresh_tokens**
```sql
CREATE TABLE refresh_tokens (
    guid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_guid UUID REFERENCES users(guid),
    token_hash VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### InfluxDB Measurement

**Measurement: telemetry**
```
Measurement: telemetry
Tags: device_sn, device_type, location_name
Fields: temperature, humidity, air_quality_index, pm25, pm10, co2, voc
Timestamp: nanosecond precision
```

---

## 3. API Specification

### 3.1 Standard Response Format

**Success Response:**
```json
{
    "success": true,
    "data": { ... },
    "message": "Operation successful"
}
```

**Error Response:**
```json
{
    "success": false,
    "error": {
        "code": "ERROR_CODE",
        "message": "Human readable message"
    }
}
```

**Paginated Response:**
```json
{
    "success": true,
    "data": [ ... ],
    "pagination": {
        "page": 1,
        "limit": 10,
        "total": 100,
        "total_pages": 10
    }
}
```

### 3.2 Authentication Endpoints

#### POST /auth/login
**Request:**
```json
{
    "email": "user@example.com",
    "password": "password123"
}
```

**Response (200):**
```json
{
    "success": true,
    "data": {
        "access_token": "eyJhbGciOiJIUzI1NiIs...",
        "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
        "expires_in": 900
    }
}
```

#### POST /auth/logout
**Headers:** `Authorization: Bearer <access_token>`

**Response (200):**
```json
{
    "success": true,
    "message": "Logged out successfully"
}
```

#### POST /auth/register
**Request:**
```json
{
    "email": "user@example.com",
    "password": "password123",
    "first_name": "John",
    "last_name": "Doe"
}
```

**Response (201):**
```json
{
    "success": true,
    "data": {
        "guid": "550e8400-e29b-41d4-a716-446655440000",
        "email": "user@example.com",
        "first_name": "John",
        "last_name": "Doe"
    }
}
```

#### POST /auth/forgot-password
**Request:**
```json
{
    "email": "user@example.com"
}
```

**Response (200):**
```json
{
    "success": true,
    "message": "Password reset instructions sent to email"
}
```

#### POST /auth/token/refresh
**Request:**
```json
{
    "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

**Response (200):**
```json
{
    "success": true,
    "data": {
        "access_token": "eyJhbGciOiJIUzI1NiIs...",
        "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
        "expires_in": 900
    }
}
```

### 3.3 Master Data - User Endpoints

#### POST /user
**Headers:** `Authorization: Bearer <access_token>`
**Request:**
```json
{
    "email": "user@example.com",
    "password": "password123",
    "first_name": "John",
    "last_name": "Doe"
}
```

**Response (201):**
```json
{
    "success": true,
    "data": {
        "guid": "550e8400-e29b-41d4-a716-446655440000",
        "email": "user@example.com",
        "first_name": "John",
        "last_name": "Doe",
        "created_at": "2026-04-29T10:00:00Z"
    }
}
```

#### GET /user/{guid}
**Headers:** `Authorization: Bearer <access_token>`

**Response (200):**
```json
{
    "success": true,
    "data": {
        "guid": "550e8400-e29b-41d4-a716-446655440000",
        "email": "user@example.com",
        "first_name": "John",
        "last_name": "Doe",
        "is_active": true,
        "created_at": "2026-04-29T10:00:00Z",
        "updated_at": "2026-04-29T10:00:00Z"
    }
}
```

#### POST /user/list
**Headers:** `Authorization: Bearer <access_token>`
**Request:**
```json
{
    "limit": 10,
    "page": 1,
    "order": "created_at",
    "sort": "DESC"
}
```

**Response (200):**
```json
{
    "success": true,
    "data": [ ... ],
    "pagination": {
        "page": 1,
        "limit": 10,
        "total": 50,
        "total_pages": 5
    }
}
```

#### PATCH /user/{guid}
**Headers:** `Authorization: Bearer <access_token>`
**Request:**
```json
{
    "email": "newemail@example.com",
    "first_name": "Jane"
}
```

#### DELETE /user/{guid}
**Headers:** `Authorization: Bearer <access_token>`

Soft delete - sets `deleted_at` timestamp.

**Response (200):**
```json
{
    "success": true,
    "message": "User deleted successfully"
}
```

### 3.4 Master Data - Device Endpoints

#### POST /device
**Headers:** `Authorization: Bearer <access_token>`
**Request:**
```json
{
    "type": "AQI-SENSOR-V1",
    "serial_number": "SN-2024-001",
    "alias": "Sensor Lantai 1",
    "notes": "Installed near window A"
}
```

#### GET /device/{guid}

#### POST /device/list
```json
{
    "limit": 10,
    "page": 1,
    "order": "created_at",
    "sort": "DESC"
}
```

#### PATCH /device/{guid}
```json
{
    "alias": "Updated Alias",
    "notes": "Updated notes"
}
```

#### DELETE /device/{guid}
Soft delete.

### 3.5 Master Data - Location Endpoints

#### POST /location
**Request:**
```json
{
    "name": "Gedung A - Lantai 1",
    "notes": "Area produksi utama"
}
```

#### GET /location/{guid}

#### POST /location/list
```json
{
    "limit": 10,
    "page": 1,
    "order": "created_at",
    "sort": "DESC"
}
```

#### PATCH /location/{guid}
```json
{
    "name": "Gedung A - Lantai 2",
    "notes": "Area pergudangan"
}
```

#### DELETE /location/{guid}
Soft delete.

### 3.6 Master Data - Installation Point Endpoints

#### POST /installation-point
**Request:**
```json
{
    "name": "Titik Monitoring A1",
    "device_guid": "550e8400-e29b-41d4-a716-446655440001",
    "location_guid": "550e8400-e29b-41d4-a716-446655440002",
    "notes": "Near air vent"
}
```

#### GET /installation-point/{guid}

#### POST /installation-point/list
```json
{
    "limit": 10,
    "page": 1,
    "order": "created_at",
    "sort": "DESC"
}
```

#### PATCH /installation-point/{guid}
```json
{
    "name": "Titik Monitoring A2",
    "device_guid": "550e8400-e29b-41d4-a716-446655440003",
    "location_guid": "550e8400-e29b-41d4-a716-446655440002"
}
```

#### DELETE /installation-point/{guid}
Soft delete.

### 3.7 Master Data - API Key Endpoints

#### POST /apikey
**Request:**
```json
{
    "notes": "API Key untuk Device IoT #1",
    "expire_date": "2027-12-31T23:59:59Z",
    "is_active": true
}
```

**Response (201):**
```json
{
    "success": true,
    "data": {
        "guid": "550e8400-e29b-41d4-a716-446655440010",
        "api_key": "aeth_live_pk_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
        "notes": "API Key untuk Device IoT #1",
        "expire_date": "2027-12-31T23:59:59Z",
        "is_active": true,
        "created_at": "2026-04-29T10:00:00Z"
    }
}
```

> ⚠️ **api_key hanya muncul sekali saat creation**, tidak bisa di-retrieve ulang.

#### GET /apikey/{guid}

#### POST /apikey/list
```json
{
    "limit": 10,
    "page": 1,
    "order": "created_at",
    "sort": "DESC"
}
```

#### PATCH /apikey/{guid}
```json
{
    "notes": "Updated notes",
    "expire_date": "2028-12-31T23:59:59Z",
    "is_active": false
}
```

#### DELETE /apikey/{guid}
Soft delete.

### 3.8 Dashboard - Realtime Endpoints

#### GET /stream
**Headers:** `Authorization: Bearer <access_token>`  
**Response:** SSE (text/event-stream)

```
event: telemetry
data: {"device_sn":"SN-2024-001","temperature":25.5,"humidity":60,"aqi":45,"timestamp":"2026-04-29T10:00:00Z"}

event: telemetry
data: {"device_sn":"SN-2024-002","temperature":26.1,"humidity":58,"aqi":52,"timestamp":"2026-04-29T10:00:01Z"}
```

#### GET /stream/{device-sn}
**Headers:** `Authorization: Bearer <access_token>`  
**Response:** SSE untuk specific device only

### 3.9 Dashboard - History Endpoints

#### POST /history/telemetry/{device-sn}
**Headers:** `Authorization: Bearer <access_token>`
**Request:**
```json
{
    "start_time": "2026-04-01T00:00:00Z",
    "end_time": "2026-04-29T23:59:59Z",
    "limit": 100,
    "page": 1,
    "order": "time",
    "sort": "DESC"
}
```

**Response (200):**
```json
{
    "success": true,
    "data": [
        {
            "device_sn": "SN-2024-001",
            "temperature": 25.5,
            "humidity": 60.0,
            "aqi": 45,
            "pm25": 12.5,
            "pm10": 23.1,
            "co2": 412,
            "voc": 0.3,
            "timestamp": "2026-04-29T10:00:00Z"
        }
    ],
    "pagination": {
        "page": 1,
        "limit": 100,
        "total": 5000,
        "total_pages": 50
    }
}
```

---

## 4. Authentication & Security

### 4.1 JWT Token Structure

**Access Token Claims:**
```json
{
    "sub": "user_guid",
    "email": "user@example.com",
    "role": "admin",
    "type": "access",
    "exp": 1745925600,
    "iat": 1745924700
}
```

**Refresh Token Claims:**
```json
{
    "sub": "user_guid",
    "type": "refresh",
    "exp": 1746529500,
    "iat": 1745924700
}
```

**Token Expiration:**
| Token Type | Expiration |
|------------|------------|
| Access Token | 15 minutes |
| Refresh Token | 7 days |

### 4.2 API Key Structure
```
aeth_live_pk_{32_random_characters}
aeth_test_pk_{32_random_characters}
```

### 4.3 Authentication Flow

**1. Login:**
```
User → POST /auth/login → JWT (access + refresh)
```

**2. Access Protected Resource:**
```
User → Header: Authorization: Bearer <access_token> → Resource
```

**3. Refresh Token:**
```
User → POST /auth/token/refresh → New JWT pair
```

**4. Device Authentication (API Key):**
```
Device → Header: X-API-Key: <api_key> → Telemetry Data
```

### 4.4 Password Requirements
- Minimum 8 characters
- bcrypt hashing dengan cost factor 10

---

## 5. Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `AUTH_INVALID_CREDENTIALS` | 401 | Email atau password salah |
| `AUTH_TOKEN_EXPIRED` | 401 | Token sudah expire |
| `AUTH_TOKEN_INVALID` | 401 | Token tidak valid |
| `AUTH_REFRESH_TOKEN_INVALID` | 401 | Refresh token tidak valid |
| `RESOURCE_NOT_FOUND` | 404 | GUID tidak ditemukan |
| `RESOURCE_DELETED` | 404 | Resource sudah di-soft delete |
| `VALIDATION_ERROR` | 400 | Request body tidak valid |
| `DUPLICATE_ENTRY` | 409 | Email/Serial number sudah ada |
| `PERMISSION_DENIED` | 403 | Tidak punya akses |
| `INTERNAL_ERROR` | 500 | Server error |

---

## 6. Middleware Pipeline

```
Request → [CORS] → [Rate Limiter] → [Auth/JWT] → [Logger] → Handler
```

### 6.1 Middleware List
1. **CORS** - Allow cross-origin requests
2. **Rate Limiter** - 100 requests/minute per IP
3. **Auth Middleware** - Validate JWT token
4. **API Key Middleware** - For device telemetry endpoints
5. **Logger** - Log all requests

---

## 7. Configuration

### config.yaml
```yaml
app:
  name: "Aether Node"
  host: "0.0.0.0"
  port: 8080
  env: "production"

database:
  host: "${POSTGRES_HOST}"
  port: 5432
  user: "${POSTGRES_USER}"
  password: "${POSTGRES_PASSWORD}"
  name: "aether_node"
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime: 300

influxdb:
  url: "${INFLUXDB_URL}"
  token: "${INFLUXDB_TOKEN}"
  org: "${INFLUXDB_ORG}"
  bucket: "telemetry"

jwt:
  secret: "${JWT_SECRET}"
  access_token_ttl: 900
  refresh_token_ttl: 604800

security:
  bcrypt_cost: 10
  rate_limit_per_minute: 100
```

---

## 8. Acceptance Criteria

### 8.1 Authentication
- [ ] User dapat login dengan email dan password yang valid
- [ ] User mendapat access token dan refresh token setelah login
- [ ] User tidak dapat akses endpoint protected tanpa token
- [ ] User dapat logout dan token di-invalidate
- [ ] User dapat refresh token sebelum expire
- [ ] User dapat register akun baru

### 8.2 Master Data CRUD
- [ ] Admin dapat membuat user, device, location, installation point, api key
- [ ] Admin dapat melihat detail resource by GUID
- [ ] Admin dapat list resource dengan pagination dan sorting
- [ ] Admin dapat update resource
- [ ] Admin dapat soft-delete resource (deleted_at ter-set)

### 8.3 Dashboard Realtime
- [ ] SSE endpoint /stream mengirim data semua device
- [ ] SSE endpoint /stream/{device-sn} mengirim data device spesifik
- [ ] Data ter-update secara realtime

### 8.4 Dashboard History
- [ ] User dapat query histori data dari InfluxDB
- [ ] Hasil query support pagination
- [ ] Data diurutkan berdasarkan timestamp

### 8.5 Non-Functional
- [ ] API response time < 200ms untuk CRUD
- [ ] Password tersimpan sebagai bcrypt hash
- [ ] Soft delete tidak menghilangkan data dari database
- [ ] API Key di-hash sebelum disimpan

---

## 9. Out of Scope v1.0

- Email/SMS notification gateway
- Push notification
- Mobile application
- Web dashboard frontend
- Multi-tenancy
- Advanced analytics / ML
- Device firmware update
- Alerting system
- Data export (CSV/Excel)

---

## 10. Dependencies

### Go Packages
```go
require (
    github.com/labstack/echo/v4 v4.11.0
    github.com/jackc/pgx/v5 v5.5.0
    github.com/influxdata/influxdb-client-go/v2 v2.12.0
    github.com/golang-jwt/jwt/v5 v5.2.0
    github.com/google/uuid v1.6.0
    golang.org/x/crypto v0.18.0
    github.com/sqlc-dev/sqlc v1.26.0
)
```

---

## 11. Revision History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 29 Apr 2026 | - | Initial SRS |
