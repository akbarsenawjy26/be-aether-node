# Aether Node — Software Requirements Specification (SRS)

**Version:** 1.0  
**Date:** 2026-05-04  
**Author:** Akbar Sena Wijaya  
**Status:** Draft  

---

## 1. Introduction

### 1.1 Purpose
This document specifies the functional and non-functional requirements for the Aether Node Dashboard — a Next.js web application that interfaces with the be-aether-node backend API.

### 1.2 Scope
Frontend web application for:
- Authentication (login, register, logout, token refresh)
- Real-time telemetry monitoring via Server-Sent Events (SSE)
- Historical telemetry data visualization
- CRUD operations for Devices, Locations, Installation Points, Users, and API Keys

### 1.3 Definitions & Acronyms

| Term | Definition |
|------|-----------|
| **GUID** | Global Unique Identifier — primary key for master data records |
| **Device SN** | Device Serial Number — unique identifier for telemetry data |
| **SSE** | Server-Sent Events — unidirectional real-time data stream |
| **JWT** | JSON Web Token — stateless authentication token |
| **Telemetry** | Time-series data emitted by IoT devices |
| **IP** | Installation Point — physical location where a device is installed |

---

## 2. System Architecture

### 2.1 Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                    Aether Node Frontend                     │
│                        (Next.js 14)                          │
│  ┌──────────┐  ┌──────────────┐  ┌───────────────────────┐ │
│  │  Pages   │  │ Components   │  │   State / Data Layer  │ │
│  │ /devices │  │ TelemetryCard│  │   Zustand + TanStack  │ │
│  │ /telemetry│ │ TelemetryChart│ │   Query + React Hook  │ │
│  │ /users   │  │ DataTable    │  │   Form + Zod          │ │
│  └────┬─────┘  └──────┬───────┘  └───────────┬───────────┘ │
│       │               │                      │             │
│       └───────────────┼──────────────────────┘             │
│                       │                                     │
│                 ┌─────▼─────┐                               │
│                 │   Axios   │                               │
│                 │  Client   │                               │
│                 └─────┬─────┘                               │
└───────────────────────┼─────────────────────────────────────┘
                        │ HTTP + SSE
                        ▼
        ┌───────────────────────────────────┐
        │        be-aether-node (Go)         │
        │          Echo Framework           │
        │  ┌─────────┐  ┌─────────────────┐ │
        │  │ REST API│  │  SSE Endpoints   │ │
        │  └────┬────┘  └────────┬─────────┘ │
        │       │               │           │
        │  ┌────▼────┐    ┌─────▼────────┐   │
        │  │PostgreSQL│   │  InfluxDB    │   │
        │  │(master)  │   │ (telemetry)  │   │
        │  └──────────┘   └──────────────┘   │
        └───────────────────────────────────┘
```

### 2.2 Frontend Directory Structure

```
frontend/
├── app/                          # Next.js App Router
│   ├── (auth)/                   # Auth group (login, register)
│   │   ├── login/page.tsx
│   │   └── register/page.tsx
│   ├── (dashboard)/              # Protected dashboard group
│   │   ├── layout.tsx           # Sidebar + header layout
│   │   ├── page.tsx            # Dashboard home
│   │   ├── devices/
│   │   │   ├── page.tsx        # Device list
│   │   │   └── [guid]/page.tsx # Device detail
│   │   ├── locations/page.tsx
│   │   ├── installation-points/
│   │   │   ├── page.tsx
│   │   │   └── [guid]/page.tsx
│   │   ├── telemetry/
│   │   │   ├── stream/page.tsx
│   │   │   └── history/[device_sn]/page.tsx
│   │   ├── users/page.tsx
│   │   ├── api-keys/page.tsx
│   │   └── settings/page.tsx
│   ├── not-found.tsx
│   └── layout.tsx               # Root layout (providers)
├── components/
│   ├── ui/                      # shadcn/ui components
│   ├── auth/                    # Auth-specific components
│   ├── dashboard/               # Dashboard-specific components
│   ├── telemetry/               # Telemetry-specific components
│   └── shared/                  # Shared components
├── lib/
│   ├── api/                     # Axios client + API functions
│   │   ├── client.ts           # Axios instance with interceptors
│   │   ├── auth.ts             # Auth API calls
│   │   ├── devices.ts
│   │   ├── locations.ts
│   │   ├── telemetry.ts
│   │   └── ...
│   ├── sse/                     # SSE client utilities
│   │   ├── sseClient.ts        # EventSource wrapper with reconnect
│   │   └── types.ts            # SSE event types
│   ├── stores/                  # Zustand stores
│   │   ├── authStore.ts
│   │   ├── themeStore.ts
│   │   └── notificationStore.ts
│   ├── hooks/                   # Custom React hooks
│   │   ├── useAuth.ts
│   │   ├── useSSE.ts
│   │   ├── useTelemetryHistory.ts
│   │   └── useDebounce.ts
│   ├── utils/                  # Utility functions
│   │   ├── cn.ts               # classname merge (clsx + twMerge)
│   │   └── format.ts           # Date, number formatting
│   └── validations/            # Zod schemas
│       ├── auth.ts
│       ├── device.ts
│       └── telemetry.ts
├── types/                      # Shared TypeScript types
│   ├── api.ts                  # API response types
│   ├── device.ts
│   ├── telemetry.ts
│   └── user.ts
└── public/
```

---

## 3. Functional Requirements

### 3.1 Authentication Module

#### FR-AUTH-001: User Login
- **Input:** email (string, required), password (string, required)
- **Process:** Call `POST /auth/login`, store JWT in httpOnly cookie
- **Output:** Redirect to dashboard on success, show error toast on failure
- **Validation:** Email format, password min 6 chars
- **UI:** Clean login form with logo, email input, password input, submit button
- **Error states:** Invalid credentials, network error, server error

#### FR-AUTH-002: User Registration
- **Input:** email, password, name (all string, required)
- **Process:** Call `POST /auth/register`
- **Output:** Success message + redirect to login
- **Validation:** Email format, password min 8 chars, name min 2 chars
- **UI:** Registration form, link to login
- **Note:** Registration may be restricted to admin only in production

#### FR-AUTH-003: User Logout
- **Process:** Call `POST /auth/logout`, clear auth state, redirect to login
- **Output:** Clean logout with no back-navigation allowed

#### FR-AUTH-004: Token Refresh
- **Trigger:** Automatic when access token expires (React Query retry)
- **Process:** Call `POST /auth/token/refresh` with refresh token
- **Output:** New access token, continue pending request
- **Edge case:** Refresh token expired → redirect to login

#### FR-AUTH-005: Protected Routes
- All dashboard routes require valid JWT
- Unauthenticated users redirected to `/login`
- Token validated via API call to `/user/me` on initial load

---

### 3.2 Dashboard Module

#### FR-DASH-001: Dashboard Overview
- **Page:** `/` (authenticated)
- **Content:**
  - Total devices count (online/offline)
  - Recent telemetry readings (last 5)
  - Quick stats (location count, user count)
  - Recent activity log
- **Refresh:** Auto-refresh every 30 seconds

#### FR-DASH-002: Sidebar Navigation
- **Items:** Dashboard, Devices, Locations, Installation Points, Telemetry (Stream, History), Users*, API Keys*, Settings
- **Active state:** Highlight current page
- **Collapse:** Toggle sidebar collapse on mobile
- **Admin-only items:** Users, API Keys (show 403 or hide based on role)

---

### 3.3 Device Management Module

#### FR-DEV-001: Device List
- **Page:** `GET /device/list` → `/devices`
- **Table columns:** Name, Serial Number, Type, Location, Status, Created At, Actions
- **Features:** Search (by name or SN), sort, pagination (10/25/50 per page)
- **Actions:** View detail, Edit, Delete (with confirmation)
- **Empty state:** "No devices found" with CTA to add

#### FR-DEV-002: Create Device
- **Page:** `/devices/new` (modal or page)
- **Fields:**
  - name (string, required, max 100)
  - serial_number (string, required, unique, max 50)
  - type (enum: sensor, gateway, controller, other)
  - location_guid (UUID, optional, FK to location)
  - metadata (JSON, optional)
- **Process:** `POST /device` → redirect to device list
- **Validation:** Zod schema on client + backend validation
- **Success:** Toast + redirect
- **Error:** Inline field errors + toast for server errors

#### FR-DEV-003: Device Detail
- **Page:** `/devices/[guid]`
- **Content:** Device info card, latest telemetry reading, linked installation points
- **Actions:** Edit, Delete, View Telemetry Stream, View History
- **404:** "Device not found" page if GUID invalid

#### FR-DEV-004: Update Device
- **Process:** `PATCH /device/:guid` with partial payload
- **Optimistic UI:** Update local state immediately, rollback on error

#### FR-DEV-005: Delete Device
- **Confirmation:** Modal with "Type device name to confirm"
- **Process:** `DELETE /device/:guid`
- **Cascade:** Warn if device has linked installation points

---

### 3.4 Location Management Module

#### FR-LOC-001: Location List
- **Page:** `POST /location/list` → `/locations`
- **Table columns:** Name, Address, Latitude, Longitude, Device Count, Actions
- **Features:** Search, sort, pagination

#### FR-LOC-002: Create/Update/Delete Location
- Similar pattern to Device CRUD
- **Fields:** name (required), address (optional), latitude (number, optional), longitude (number, optional)

---

### 3.5 Installation Point Module

#### FR-IP-001: Installation Point List
- **Page:** `POST /installation-point/list` → `/installation-points`
- **Table columns:** Name, Device SN, Location Name, Installed At, Actions
- **Features:** Search, sort, pagination, filter by location

#### FR-IP-002: Get IP with Relations
- **Endpoint:** `GET /installation-point/:guid/relations`
- **Use:** Show device info + location info in single view

#### FR-IP-003: Create/Update/Delete Installation Point
- **Fields:** name, device_guid, location_guid, installed_at (datetime), notes (optional)

---

### 3.6 User Management Module

#### FR-USER-001: User List
- **Page:** `POST /user/list` → `/users` (Admin only)
- **Table columns:** Name, Email, Role, Created At, Last Login, Actions
- **Actions:** View, Edit, Delete

#### FR-USER-002: Create User
- **Fields:** name, email, password, role (enum: admin, operator, viewer)

#### FR-USER-003: Update User
- **Fields:** name, email, role, password (optional, only if changing)
- **Self-edit:** Users can edit their own profile at `/settings`

---

### 3.7 API Key Management Module

#### FR-APIKEY-001: API Key List
- **Page:** `POST /apikey/list` → `/api-keys` (Admin only)
- **Table columns:** Name, Key (masked: `ak_xxxx...xxxx`), Device, Created At, Expires At, Status, Actions
- **Note:** Full key only shown once at creation

#### FR-APIKEY-002: Create API Key
- **Fields:** name (required), device_guid (optional), expires_at (datetime, optional)
- **Process:** `POST /apikey` → display full key ONCE with copy button
- **Warning:** "This key will not be shown again"

#### FR-APIKEY-003: Revoke API Key
- **Process:** `DELETE /apikey/:guid`
- **Confirmation:** Modal "Revoke this key? Devices using it will lose access."

---

### 3.8 Telemetry Stream Module (SSE)

#### FR-TELEM-STREAM-001: SSE Connection Manager
- **Connection:** `GET /telemetry/stream?token=<jwt>` or via cookie
- **Events:** `device_data`, `connected`, `error`, `heartbeat`
- **Event format:**
  ```json
  {
    "event": "device_data",
    "data": {
      "device_sn": "SN001",
      "device_type": "sensor",
      "readings": {
        "temperature": 25.4,
        "humidity": 60.2
      },
      "timestamp": "2026-05-04T20:08:00Z"
    }
  }
  ```
- **Reconnect logic:** On disconnect, wait 3s, retry. Max 5 retries with exponential backoff.
- **UI states:** Connecting, Connected (green dot), Disconnected (red dot + banner), Reconnecting

#### FR-TELEM-STREAM-002: Stream All Devices
- **Page:** `/telemetry/stream`
- **Layout:** Grid of `TelemetryCard` components (auto-arrange 1-4 columns)
- **Each card shows:**
  - Device SN + name
  - Latest readings (key-value pairs)
  - Timestamp of last update
  - Status indicator (active/stale: >30s no update = stale)
  - Mini sparkline (optional)
- **Controls:** Pause stream, Filter by device type, Filter by location

#### FR-TELEM-STREAM-003: Stream Single Device
- **Page:** `/telemetry/stream/[device_sn]`
- **Layout:** Larger card with full readings + scrollable event log
- **Actions:** Start/Stop stream, copy device SN, go to history

---

### 3.9 Telemetry History Module

#### FR-TELEM-HIST-001: Historical Query
- **Page:** `/telemetry/history/[device_sn]`
- **API:** `POST /telemetry/history/:device_sn`
- **Request body:**
  ```json
  {
    "start": "2026-05-04T00:00:00Z",
    "stop": "2026-05-04T23:59:59Z",
    "limit": 100,
    "order": "desc",
    "window": "1m"
  }
  ```
- **Default:** Last 24 hours, 100 records, descending order

#### FR-TELEM-HIST-002: Chart Visualization
- **Library:** Recharts
- **Chart type:** Line chart (time on X-axis, readings on Y-axis)
- **Features:**
  - Toggle series visibility
  - Zoom (brush component)
  - Tooltip with exact values
  - Responsive container
- **Time windows:** 1m, 5m, 15m, 1h (aggregation interval)
- **Date range picker:** Preset buttons (Last 1h, 6h, 24h, 7d, 30d) + custom range

#### FR-TELEM-HIST-003: Data Table
- **Below chart:** Sortable, paginated table of raw data
- **Columns:** Timestamp, device_sn, all reading fields
- **Export:** Download as CSV button

---

## 4. Non-Functional Requirements

### 4.1 Performance
- **First Contentful Paint:** < 1.5s
- **Time to Interactive:** < 3s
- **SSE latency:** < 2s from backend to UI update
- **API response (CRUD):** < 500ms
- **History query:** < 5s for 1000 records

### 4.2 Reliability
- **SSE reconnection:** Automatic with exponential backoff (3s → 6s → 12s → 24s → 48s, max 5 retries)
- **API retry:** TanStack Query retry 3 times on GET, no retry on mutations
- **Offline detection:** Show banner when network lost

### 4.3 Security
- **JWT storage:** httpOnly cookie (not localStorage)
- **HTTPS:** All production traffic over HTTPS
- **CORS:** Backend validates Origin header
- **Input sanitization:** Zod validation on all inputs
- **XSS prevention:** React auto-escapes, no dangerouslySetInnerHTML
- **CSRF:** SameSite=Strict cookie + CSRF token for state-changing operations

### 4.4 Accessibility
- **Keyboard navigation:** All interactive elements focusable
- **ARIA labels:** On icons, buttons without text
- **Color contrast:** WCAG AA compliant
- **Screen reader:** Semantic HTML

### 4.5 Browser Support
- Chrome 90+
- Firefox 90+
- Safari 14+
- Edge 90+

---

## 5. Data Models

### 5.1 API Response Wrapper

```typescript
// Success
interface ApiResponse<T> {
  success: true;
  data: T;
  message?: string;
}

// Error
interface ApiError {
  success: false;
  error: {
    code: string;
    message: string;
    details?: Record<string, string[]>;
  };
}
```

### 5.2 Domain Types

```typescript
// Device
interface Device {
  guid: string;
  name: string;
  serial_number: string;
  type: 'sensor' | 'gateway' | 'controller' | 'other';
  location_guid?: string;
  location_name?: string;
  status: 'online' | 'offline' | 'unknown';
  metadata?: Record<string, unknown>;
  created_at: string;
  updated_at: string;
}

// Location
interface Location {
  guid: string;
  name: string;
  address?: string;
  latitude?: number;
  longitude?: number;
  device_count?: number;
  created_at: string;
  updated_at: string;
}

// Installation Point
interface InstallationPoint {
  guid: string;
  name: string;
  device_guid?: string;
  device_sn?: string;
  location_guid?: string;
  location_name?: string;
  installed_at?: string;
  notes?: string;
  created_at: string;
  updated_at: string;
}

// User
interface User {
  guid: string;
  name: string;
  email: string;
  role: 'admin' | 'operator' | 'viewer';
  created_at: string;
  updated_at: string;
}

// Telemetry Reading
interface TelemetryReading {
  device_sn: string;
  device_type?: string;
  device_name?: string;
  readings: Record<string, number>;
  timestamp: string;
}

// History Response
interface HistoryResponse {
  device_sn: string;
  columns: string[];
  values: (string | number)[][];
  row_count: number;
}

// API Key
interface APIKey {
  guid: string;
  name: string;
  key_masked: string;
  device_guid?: string;
  device_sn?: string;
  expires_at?: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}
```

---

## 6. SSE Event Types

```typescript
// SSE event types
type SSEEventType = 'connected' | 'device_data' | 'error' | 'heartbeat';

interface SSEDeviceData {
  device_sn: string;
  device_type: string;
  device_name?: string;
  readings: Record<string, number>;
  timestamp: string;
}

interface SSEError {
  code: string;
  message: string;
}

// EventSource message format
interface SSEMessage {
  event: SSEEventType;
  data: SSEDeviceData | SSEError | { count: number };
  timestamp?: string;
}
```

---

## 7. Error Handling

### 7.1 Error Display Strategy

| Error Type | UI Response |
|-----------|-------------|
| Validation error (400) | Inline field errors (red border + message below field) |
| Unauthorized (401) | Redirect to login + toast "Session expired" |
| Forbidden (403) | Toast + stay on page (or 403 page) |
| Not found (404) | "Resource not found" page |
| Server error (500) | Toast "Something went wrong. Please try again." |
| Network error | Toast "Network error. Retrying..." + auto-retry |

### 7.2 Toast Notifications
- Position: Bottom-right
- Types: success (green), error (red), warning (yellow), info (blue)
- Duration: 5s auto-dismiss, hover to pause
- Max visible: 3 (older ones queue)

---

## 8. Appendix

### 8.1 Environment Variables

```env
NEXT_PUBLIC_API_BASE_URL=http://localhost:8080
NEXT_PUBLIC_APP_NAME=Aether Node
```

### 8.2 Mock Data Strategy
- Use MSW (Mock Service Worker) for development
- Mock endpoints for all API calls
- Mock SSE using `EventSource` polyfill or custom mock stream

### 8.3 Testing Strategy
- **Unit tests:** Vitest + React Testing Library
- **Component tests:** Storybook + Chromatic
- **E2E tests:** Playwright
- **Coverage target:** 70%+

### 8.4 Deployment
- **Hosting:** Vercel (recommended for Next.js)
- **CI/CD:** GitHub Actions
- **Environments:** Development, Staging, Production
