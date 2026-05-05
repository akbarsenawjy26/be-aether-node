# Aether Node — Product Requirements Document (PRD)

**Version:** 1.0  
**Date:** 2026-05-04  
**Author:** Akbar Sena Wijaya  
**Status:** Draft  

---

## 1. Overview

### 1.1 Product Name
**Aether Node Dashboard** — IoT Telemetry Monitoring Platform

### 1.2 Product Summary
Web dashboard untuk monitoring data telemetry real-time dari device IoT. Frontend Next.js yang terhubung ke backend be-aether-node (Go/Echo), menampilkan data stream real-time (SSE), historical data (chart), dan CRUD untuk master data (device, location, installation point, user, api key).

### 1.3 Target Users

| Role | Description |
|------|-------------|
| **Admin** | Full access: manage users, devices, locations, installation points, API keys |
| **Operator** | Monitor real-time telemetry, view historical data, manage devices |
| **Viewer** | Read-only dashboard view, no management capabilities |

---

## 2. Problem Statement

- Device IoT mengirim data telemetry secara terus-menerus (SSE stream)
- Tim operations perlu melihat data real-time tanpa harus pake curl/Postman
- Historical data perlu divisualisasikan dalam bentuk chart
- Master data (device, location, installation point) perlu management yang mudah
- Belum ada UI untuk mengelola user dan API keys

---

## 3. Goals

### 3.1 Primary Goals
- [ ] Real-time telemetry dashboard dengan SSE stream
- [ ] Historical data visualization (line chart, time-series)
- [ ] Device management (CRUD + status)
- [ ] Location management (CRUD + geo info)
- [ ] Installation point management (CRUD + relations)
- [ ] User management (CRUD)
- [ ] API key management (create, list, revoke)
- [ ] Authentication (login, register, logout, JWT)

### 3.2 Secondary Goals
- [ ] Dark/Light theme toggle
- [ ] Responsive design (mobile-friendly)
- [ ] Notifications (new device, connection lost, threshold alert)
- [ ] Export historical data (CSV)

---

## 4. User Stories

### 4.1 Authentication
```
As a user,
I want to login with email/password,
So that I can access the dashboard securely.
```

```
As a user,
I want to logout,
So that my session is terminated.
```

```
As an admin,
I want to register new users,
So that new team members can access the system.
```

### 4.2 Device Management
```
As an admin,
I want to add new devices,
So that I can start receiving telemetry from them.
```

```
As an operator,
I want to see all devices with their status,
So that I can monitor which ones are online/offline.
```

```
As an admin,
I want to update/delete devices,
So that I can manage the device inventory.
```

### 4.3 Real-time Monitoring
```
As an operator,
I want to see live telemetry stream,
So that I can monitor device data in real-time without refreshing.
```

```
As an operator,
I want to filter stream by device,
So that I can focus on specific devices.
```

```
As an operator,
I want to see SSE connection status,
So that I know if the stream is active or disconnected.
```

### 4.4 Historical Data
```
As an operator,
I want to query historical telemetry by date range,
So that I can analyze past performance.
```

```
As an operator,
I want to see historical data as a line chart,
So that trends are easy to understand.
```

```
As an operator,
I want to paginate and sort historical data,
So that I can navigate large datasets.
```

### 4.5 Location & Installation Point
```
As an admin,
I want to manage locations,
So that I know where each device is installed.
```

```
As an admin,
I want to manage installation points,
So that I can map devices to specific locations.
```

### 4.6 API Key Management
```
As an admin,
I want to create API keys for devices,
So that devices can authenticate when sending telemetry.
```

```
As an admin,
I want to revoke API keys,
So that I can disable compromised or unused keys.
```

---

## 5. Feature List

### 5.1 Pages & Routes

| Page | Route | Access | Description |
|------|-------|--------|-------------|
| Login | `/login` | Public | Email + password login |
| Register | `/register` | Public | New user registration |
| Forgot Password | `/forgot-password` | Public | Request password reset |
| Dashboard / Realtime | `/dashboard/realtime` | Auth | Live SSE stream (all devices grid) |
| Dashboard / History | `/dashboard/history/[device_sn]` | Auth | Per-device historical chart + table |
| Master Data / Users | `/master-data/users` | Admin | User CRUD list |
| Master Data / Devices | `/master-data/devices` | Auth | Device list + CRUD |
| Master Data / Devices Detail | `/master-data/devices/[guid]` | Auth | Single device detail |
| Master Data / Locations | `/master-data/locations` | Auth | Location list + CRUD |
| Master Data / Installation Points | `/master-data/installation-points` | Auth | Installation point list + CRUD |
| Master Data / API Keys | `/master-data/api-keys` | Admin | API key create + revoke |
| Settings / Profile | `/settings` | Auth | Profile edit |
| Settings / Change Password | `/settings/password` | Auth | Change password |
| 404 | `/*` | Public | Not found page |

### 5.2 Component Library

| Component | Description |
|-----------|-------------|
| `TelemetryCard` | Card displaying latest reading from a device |
| `TelemetryChart` | Line chart for historical data (Recharts) |
| `SSEReader` | SSE connection manager with auto-reconnect |
| `DataTable` | Sortable, paginated table (TanStack Table) |
| `DeviceStatusBadge` | Online/offline/unknown indicator |
| `LocationMap` | Optional: map view of locations |
| `StatCard` | Dashboard KPI card |
| `Modal` | Reusable modal dialog |
| `Toast` | Notification toaster |
| `Sidebar` | Navigation sidebar |
| `ThemeToggle` | Dark/light mode switch |

---

## 6. Technical Stack

| Layer | Technology |
|-------|-----------|
| Framework | Next.js 14 (App Router) |
| Language | TypeScript |
| Styling | Tailwind CSS |
| UI Components | shadcn/ui |
| State Management | Zustand |
| Data Fetching | TanStack Query (React Query) |
| Charts | Recharts |
| Tables | TanStack Table |
| Forms | React Hook Form + Zod |
| HTTP Client | Axios |
| SSE Client | Native EventSource + custom reconnect logic |
| Auth | JWT stored in httpOnly cookie |
| Icons | Lucide React |

### 6.1 Branding

| Element | Value |
|---------|-------|
| Primary Color | `#517E68` (sage green) |
| White | `#FFFFFF` |
| Dark Mode | Default dark theme |
| Language | Indonesian (id-ID) |

### 6.2 Menu Structure

```
📌 SIDEBAR NAVIGATION

1. Auth
   └── Login
   └── Register
   └── Forgot Password

2. Dashboard
   └── Realtime (SSE stream all devices)
   └── History (per-device historical chart)

3. Master Data (sub-menu expandable)
   └── User Management
   └── Device Management
   └── Location Management
   └── Installation Point Management
   └── API Key Management

4. Settings
   └── Profile
   └── Change Password
```

---

## 7. API Integration

### 7.1 Backend Base URL
- **Local:** `http://localhost:8080`
- **VPS:** `http://103.127.132.230:8080`

### 7.2 Authentication Flow
```
1. User submits login form
2. Frontend calls POST /auth/login
3. Backend returns JWT access_token + refresh_token
4. Frontend stores tokens (httpOnly cookie atau secure storage)
5. Subsequent requests include Authorization: Bearer <token>
6. Token expired → React Query retry → call POST /auth/token/refresh
```

### 7.3 SSE Integration
```
1. Frontend connects to GET /telemetry/stream (with JWT query param)
2. Backend streams SSE events: device_sn, readings, timestamp
3. Frontend uses EventSource API with custom reconnect logic
4. Display updates in real-time via TelemetryCard components
5. Connection lost → show warning + auto-reconnect
```

### 7.4 API Endpoints Summary

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| POST | `/auth/login` | No | Login |
| POST | `/auth/register` | No | Register |
| POST | `/auth/token/refresh` | Refresh | Refresh token |
| POST | `/auth/logout` | Yes | Logout |
| GET | `/user/me` | Yes | Current user profile |
| GET | `/user/:guid` | Yes | Get user by GUID |
| POST | `/user/list` | Yes | List users |
| POST | `/user` | Yes | Create user |
| PATCH | `/user/:guid` | Yes | Update user |
| DELETE | `/user/:guid` | Yes | Delete user |
| GET | `/device/:guid` | Yes | Get device |
| POST | `/device/list` | Yes | List devices |
| POST | `/device` | Yes | Create device |
| PATCH | `/device/:guid` | Yes | Update device |
| DELETE | `/device/:guid` | Yes | Delete device |
| GET | `/location/:guid` | Yes | Get location |
| POST | `/location/list` | Yes | List locations |
| POST | `/location` | Yes | Create location |
| PATCH | `/location/:guid` | Yes | Update location |
| DELETE | `/location/:guid` | Yes | Delete location |
| GET | `/installation-point/:guid` | Yes | Get IP |
| GET | `/installation-point/:guid/relations` | Yes | Get IP with relations |
| POST | `/installation-point/list` | Yes | List IPs |
| POST | `/installation-point` | Yes | Create IP |
| PATCH | `/installation-point/:guid` | Yes | Update IP |
| DELETE | `/installation-point/:guid` | Yes | Delete IP |
| GET | `/apikey/:guid` | Yes | Get API key |
| POST | `/apikey/list` | Yes | List API keys |
| POST | `/apikey` | Yes | Create API key |
| PATCH | `/apikey/:guid` | Yes | Update API key |
| DELETE | `/apikey/:guid` | Yes | Delete API key |
| POST | `/telemetry` | API Key | Device data ingestion |
| GET | `/telemetry/stream` | Yes | SSE all devices |
| GET | `/telemetry/stream/:device_sn` | Yes | SSE per device |
| POST | `/telemetry/history/:device_sn` | Yes | Historical data |

---

## 8. Out of Scope (v1)

- Mobile native app
- Push notifications (mobile)
- Multi-tenant / organization
- Data export to PDF
- Advanced analytics / ML
- Device firmware update
- WebSocket (SSE is sufficient for now)
- LDAP / SSO integration
- Audit logging

---

## 9. Success Metrics (v1)

- All CRUD operations functional
- SSE stream latency < 2 seconds
- Historical query response < 5 seconds (InfluxDB)
- Zero authentication bypass vulnerabilities
- Responsive on desktop and tablet
- Zero console errors in production

---

## 10. Open Questions

1. **Dark mode** — Default dark atau light? Atau user preference stored di localStorage?
2. **Number of devices** — Expected scale? Affects pagination design.
3. **Chart library** — Recharts OK atau prefer Chart.js?
4. **File storage** — Need image upload untuk device photos?
5. **Multi-language** — Indonesian only atau English juga?
6. **Branding** — Ada logo/color scheme yang sudah固定的?
