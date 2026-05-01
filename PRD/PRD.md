# PRD - Aether Node

## Product Requirements Document
**Versi:** 1.0  
**Tanggal:** 29 April 2026  
**Produk:** Aether Node - Air Quality Management & Monitoring IoT Device

---

## 1. Overview

### 1.1 Nama Produk
**Aether Node**

### 1.2 Tagline / Summary
Dashboard untuk monitoring kualitas udara dan management device IoT secara realtime.

### 1.3 Purpose
Menyediakan platform backend yang memungkinkan industri, office, dan pergudangan untuk memantau kualitas udara secara realtime, mengelola device IoT, dan menghasilkan laporan histori data.

### 1.4 Target Users
| Segmen | Use Case |
|--------|----------|
| **Industri** | Monitoring kualitas udara di area produksi |
| **Office** | Monitoring kualitas udara di gedung perkantoran |
| **Pergudangan** | Monitoring kualitas udara di warehouse |

---

## 2. Goals

### 2.1 Business Goals
- Menyediakan sistem backend yang scalable untuk monitoring IoT device
- Mendukung integrasi dengan multiple device melalui API Key
- Menyediakan data realtime dan histori untuk analisis kualitas udara

### 2.2 Technical Goals
- Backend menggunakan Go dengan Echo Framework
- Pattern SoA (Service-Oriented Architecture) untuk scalability
- Dual database: PostgreSQL (transactional) + InfluxDB (time-series read)
- Security dengan JWT (Access Token + Refresh Token) dan API Key

---

## 3. User Stories

### 3.1 Authentication Module
| ID | User Story | Acceptance Criteria |
|----|-----------|---------------------|
| US-001 |Sebagai user, saya mau login dengan email dan password | Login berhasil → dapat access token & refresh token |
| US-002 |Sebagai user, saya mau logout | Session di-invalidate, token tidak bisa dipakai |
| US-003 |Sebagai user, saya mau register akun baru | Akun terbuat → dapat login langsung |
| US-004 |Sebagai user, saya mau reset password via email | Email terkirim → user bisa set password baru |
| US-005 |Sebagai user, saya mau refresh token | Token expired → dapat token baru tanpa login ulang |

### 3.2 Master Data - User
| ID | User Story | Acceptance Criteria |
|----|-----------|---------------------|
| US-010 |Sebagai admin, saya mau CRUD user | User di-create, read, update, soft-delete |
| US-011 |Sebagai admin, saya mau assign role ke user | User memiliki role yang sesuai |
| US-012 |Sebagai admin, saya mau lihat list user dengan pagination | List user + total count, sorting, filtering |

### 3.3 Master Data - Role Access
| ID | User Story | Acceptance Criteria |
|----|-----------|---------------------|
| US-020 |Sebagai admin, saya mau CRUD role | Role di-create, read, update, soft-delete |
| US-021 |Sebagai admin, saya mau set permission per role | Role memiliki list permissions yang sesuai |

### 3.4 Master Data - Device
| ID | User Story | Acceptance Criteria |
|----|-----------|---------------------|
| US-030 |Sebagai admin, saya mau CRUD device IoT | Device di-create, read, update, soft-delete |
| US-031 |Sebagai admin, saya mau register device baru | Device terdaftar dengan serial number unik |
| US-032 |Sebagai admin, saya mau lihat list device | List device + pagination, sorting |

### 3.5 Master Data - Location
| ID | User Story | Acceptance Criteria |
|----|-----------|---------------------|
| US-040 |Sebagai admin, saya mau CRUD location | Location di-create, read, update, soft-delete |
| US-041 |Sebagai admin, saya mau asign device ke location | Device terikat dengan installation point |

### 3.6 Master Data - Installation Point
| ID | User Story | Acceptance Criteria |
|----|-----------|---------------------|
| US-050 |Sebagai admin, saya mau CRUD installation point | Installation point di-create, read, update, soft-delete |
| US-051 |Sebagai admin, saya mau kaitkan device + location | Installation point = device + location + metadata |

### 3.7 Master Data - API Key
| ID | User Story | Acceptance Criteria |
|----|-----------|---------------------|
| US-060 |Sebagai admin, saya mau generate API key | API key ter-generate untuk device integration |
| US-061 |Sebagai admin, saya mau deactivate API key | API key tidak bisa dipakai tanpa dihapus |
| US-062 |Sebagai admin, saya mau set expire date API key | API key auto-expired sesuai tanggal |

### 3.8 Dashboard Realtime
| ID | User Story | Acceptance Criteria |
|----|-----------|---------------------|
| US-070 |Sebagai user, saya mau lihat data realtime semua device | SSE stream memberikan data terbaru |
| US-071 |Sebagai user, saya mau filter data per device | SSE stream per serial number |
| US-072 |Sebagai user, saya mau lihat live dashboard | Data ter-update otomatis tanpa refresh |

### 3.9 Dashboard History & Report
| ID | User Story | Acceptance Criteria |
|----|-----------|---------------------|
| US-080 |Sebagai user, saya mau lihat histori data device | Data diambil dari InfluxDB |
| US-081 |Sebagai user, saya mau export laporan | Data bisa di-export (format: CSV/Excel) |

---

## 4. Functional Requirements

### 4.1 Authentication
- [ ] Login dengan email + password → JWT (access + refresh token)
- [ ] Register dengan email + password + first_name + last_name
- [ ] Logout → invalidate token
- [ ] Forgot password → kirim email reset link
- [ ] Refresh token → issue token baru

### 4.2 Master Data CRUD (Soft Delete Pattern)

#### User
- [ ] Create user: email, password (hashed), first_name, last_name
- [ ] Read user by GUID
- [ ] List user: pagination (limit, page), sorting (order, sort), search
- [ ] Update user: email, password, first_name, last_name
- [ ] Delete user: soft delete (set deleted_at)

#### Role Access
- [ ] Create role: name, permissions
- [ ] Read role by GUID
- [ ] List role: pagination, sorting
- [ ] Update role: name, permissions
- [ ] Delete role: soft delete

#### Device
- [ ] Create device: type, serial_number, alias, notes
- [ ] Read device by GUID
- [ ] List device: pagination, sorting
- [ ] Update device: type, serial_number, alias, notes
- [ ] Delete device: soft delete

#### Location
- [ ] Create location: name, notes
- [ ] Read location by GUID
- [ ] List location: pagination, sorting
- [ ] Update location: name, notes
- [ ] Delete location: soft delete

#### Installation Point
- [ ] Create installation point: name, device_guid, location_guid, notes
- [ ] Read installation point by GUID
- [ ] List installation point: pagination, sorting
- [ ] Update installation point: name, device_guid, location_guid, notes
- [ ] Delete installation point: soft delete

#### API Key
- [ ] Create API key: notes, expire_date, is_active
- [ ] Read API key by GUID
- [ ] List API key: pagination, sorting
- [ ] Update API key: notes, expire_date, is_active
- [ ] Delete API key: soft delete

### 4.3 Dashboard Realtime (SSE - Server-Sent Events)
- [ ] SSE endpoint: GET /stream → all device data
- [ ] SSE endpoint: GET /stream/{device-sn} → specific device data

### 4.4 Dashboard History
- [ ] POST /history/telemetry/{device-sn} → query InfluxDB
- [ ] Support pagination, sorting
- [ ] Return telemetry data + metadata

---

## 5. Non-Functional Requirements

### 5.1 Performance
- API response time < 200ms untuk CRUD operations
- SSE latency < 500ms untuk realtime data
- Support 100+ concurrent SSE connections

### 5.2 Security
- Password hashed dengan bcrypt
- JWT dengan expiration (access: 15min, refresh: 7 days)
- API Key untuk device authentication
- Soft delete untuk data retention

### 5.3 Scalability
- SoA pattern untuk modular services
- Database connection pooling
- Stateless application design

### 5.4 Reliability
- PostgreSQL untuk persistent data (CRUD master data)
- InfluxDB untuk time-series data (telemetry)
- Graceful shutdown handling

---

## 6. API Overview

### Base URL
```
Production: https://aether-be.wit-sby.id
```

### Authentication Endpoints
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | /auth/login | User login |
| POST | /auth/logout | User logout |
| POST | /auth/register | Register new user |
| POST | /auth/forgot-password | Request password reset |
| POST | /auth/token/refresh | Refresh access token |

### Master Data Endpoints
| Module | Create | Read | List | Update | Delete |
|--------|--------|------|------|--------|--------|
| User | POST /user | GET /user/{guid} | POST /user/list | PATCH /user/{guid} | DELETE /user/{guid} |
| Role | POST /role | GET /role/{guid} | POST /role/list | PATCH /role/{guid} | DELETE /role/{guid} |
| Device | POST /device | GET /device/{guid} | POST /device/list | PATCH /device/{guid} | DELETE /device/{guid} |
| Location | POST /location | GET /location/{guid} | POST /location/list | PATCH /location/{guid} | DELETE /location/{guid} |
| Installation Point | POST /installation-point | GET /installation-point/{guid} | POST /installation-point/list | PATCH /installation-point/{guid} | DELETE /installation-point/{guid} |
| API Key | POST /apikey | GET /apikey/{guid} | POST /apikey/list | PATCH /apikey/{guid} | DELETE /apikey/{guid} |

### Dashboard Endpoints
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | /stream | SSE - All device realtime data |
| GET | /stream/{device-sn} | SSE - Specific device data |
| POST | /history/telemetry/{device-sn} | Query device history from InfluxDB |

---

## 7. Out of Scope (v1.0)

- Mobile app / Web frontend
- Push notification
- Email/SMS gateway integration (forgot password)
- Multi-tenancy
- Advanced analytics / AI predictions
- Device firmware update
- Real-time alerting / threshold rules

---

## 8. Dependencies

### External Services
- PostgreSQL (hosted)
- InfluxDB (hosted)
- Email Service Provider (future)

### Internal Dependencies
- Go 1.22+
- Echo Framework v4
- SQLC for SQL generation
- JWT library
- bcrypt for password hashing

---

## 9. Revision History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 29 Apr 2026 | - | Initial PRD |
