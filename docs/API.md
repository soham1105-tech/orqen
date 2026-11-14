# ORQEN Control Plane API Reference

ORQEN embeds a local HTTP and Server-Sent Events (SSE) control plane server accessible via `orqen serve`. This API powers the interactive web console, local integrations, and automation scripts.

---

## Server Overview

- **Default Endpoint**: `http://127.0.0.1:8080`
- **Security Headers**:
  - `Content-Security-Policy`: Restricts scripts, styles, and frames to origin
  - `X-Content-Type-Options: nosniff`
  - `X-Frame-Options: DENY`
  - `Referrer-Policy: no-referrer`

---

## Endpoints

### 1. System Health

#### `GET /api/health`

Returns server status and the build-injected ORQEN version.

**Response (`200 OK`):**
```json
{
  "ok": true,
  "version": "0.5.0"
}
```

---

### 2. Projects

#### `GET /api/projects`

Returns the list of all onboarded projects, their container states, service counts, and active URLs.

**Response (`200 OK`):**
```json
[
  {
    "id": "prj_9b2d8c3e",
    "name": "my-web-app",
    "services": 2,
    "infra": 1,
    "running": 3,
    "total": 3,
    "healthy": true,
    "urls": [
      "http://localhost:3000"
    ],
    "last_init": "2026-11-14T06:12:00Z"
  }
]
```

---

### 3. Global Configuration

#### `GET /api/config`

Returns sanitized and redacted configuration information (API keys and tokens are strictly omitted).

**Response (`200 OK`):**
```json
{
  "provider": "gemini",
  "model": "gemini-3.5-flash-lite",
  "source": "config.json",
  "key_configured": true,
  "home": "/home/user/.orqen"
}
```

---

### 4. Conversations & Chats

#### `GET /api/chats`

Lists all stored conversations sorted by most recently updated.

**Response (`200 OK`):**
```json
[
  {
    "id": "chat_a1b2c3d4",
    "project": "my-web-app",
    "project_id": "prj_9b2d8c3e",
    "title": "Diagnose api latency",
    "created_at": "2026-11-14T06:15:00Z",
    "updated_at": "2026-11-14T06:16:30Z",
    "messages": 4
  }
]
```

#### `POST /api/chats`

Creates a new conversation session associated with a project.

**Request Body:**
```json
{
  "project": "my-web-app",
  "project_id": "prj_9b2d8c3e"
}
```

**Response (`201 Created`):** Returns the full new `Chat` object.

#### `GET /api/chats/{id}`

Retrieves the complete message history and telemetry events for a conversation session.

**Response (`200 OK`):**
```json
{
  "id": "chat_a1b2c3d4",
  "project": "my-web-app",
  "project_id": "prj_9b2d8c3e",
  "title": "Diagnose api latency",
  "created_at": "2026-11-14T06:15:00Z",
  "updated_at": "2026-11-14T06:16:30Z",
  "messages": [
    {
      "id": "msg_001",
      "role": "user",
      "text": "why is the stack failing?",
      "created_at": "2026-11-14T06:15:01Z"
    },
    {
      "id": "msg_002",
      "role": "agent",
      "text": "The PostgreSQL container is failing its healthcheck due to missing POSTGRES_PASSWORD.",
      "created_at": "2026-11-14T06:15:05Z",
      "events": [
        {
          "type": "tool_call",
          "tool": "inspect_logs_and_health",
          "args": {"service": "postgres"}
        }
      ]
    }
  ]
}
```

#### `DELETE /api/chats/{id}`

Deletes a conversation from persistent storage.

**Response (`204 No Content`)**

---

### 5. Conversational Agent Streaming Turn

#### `POST /api/chat`

Executes a turn with the ORQEN DevOps agent. Returns an HTTP `text/event-stream` Server-Sent Events (SSE) connection that streams progress, tool calls, confirmation requests, and final responses in real time.

**Request Body:**
```json
{
  "chat_id": "chat_a1b2c3d4",
  "project": "my-web-app",
  "text": "restart the postgres database",
  "confirm": false,
  "retry": false
}
```

**Streamed Event Format:**

Each SSE frame is formatted as `data: { ... }\n\n`:

1. **`start`**: Emitted when the turn begins, providing chat summary metadata.
   ```json
   {
     "type": "start",
     "chat": { "id": "chat_a1b2c3d4", "project": "my-web-app" }
   }
   ```

2. **`event`**: Emitted when the agent performs an action or executes a tool.
   ```json
   {
     "type": "event",
     "event": {
       "type": "tool_call",
       "tool": "inspect_logs_and_health",
       "args": { "service": "postgres" },
       "duration_ms": 142
     }
   }
   ```

3. **`confirm`**: Emitted when a tool requires explicit user authorization before execution.
   ```json
   {
     "type": "confirm",
     "needs_confirmation": true,
     "confirmation": {
       "tool": "orchestrate_stack",
       "destructive": false,
       "args": { "action": "restart", "service": "postgres" }
     }
   }
   ```

4. **`done`**: Emitted upon completion of the agent turn with the final assistant response.
   ```json
   {
     "type": "done",
     "message": {
       "id": "msg_003",
       "role": "agent",
       "text": "PostgreSQL has been restarted and is reporting healthy.",
       "created_at": "2026-11-14T06:16:32Z"
     }
   }
   ```

5. **`error`**: Emitted if an unrecoverable server or provider error occurs.
   ```json
   {
     "type": "error",
     "error": "rate limit exceeded from LLM provider"
   }
   ```

---

## Static Assets

All web UI assets (HTML, CSS, JavaScript, SVG icons) are embedded directly into the Go binary at compile time via `go:embed`.

- `GET /` — Serves `index.html` (Single Page Application)
- `GET /assets/app.css` — Control plane styling
- `GET /assets/app.js` — Frontend state machine, SSE consumer, and event rendering
- `GET /assets/favicon.svg` — Brand mark favicon
- `GET /assets/logo.svg` — Brand logo asset
