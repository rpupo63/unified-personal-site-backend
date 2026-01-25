# Personal Site Backend API

Backend API service for managing personal site projects and blog posts.

## Database Configuration

The backend **only supports direct database connections** (port 5432). Transaction pooler connections (port 6543) are not supported.

### Required Environment Variables

You can configure the database connection using one of the following methods:

#### Option 1: Full Connection String (Recommended)

Set one of these environment variables with a complete PostgreSQL connection string:

- `DATABASE_URL` - Full PostgreSQL connection string (priority 1)
- `SUPABASE_DB_URL` - Alternative full PostgreSQL connection string (priority 2)

**Format examples:**

```
DATABASE_URL=postgres://user:password@host:5432/dbname?sslmode=require
```

**Important:** The connection string must use port **5432** (direct connection). Port 6543 (transaction pooler) is not supported and will be rejected.

#### Option 2: Individual Components

If you don't provide a full connection string, you can set individual components:

**Required:**

- `SUPABASE_DB_HOST` - Database host address
- `SUPABASE_DB_USER` - Database username
- `SUPABASE_DB_PASSWORD` - Database password

**Optional:**

- `SUPABASE_DB_NAME` - Database name (defaults to `postgres`)
- `SUPABASE_DB_PORT` - Database port (defaults to `5432`, must be `5432`)

**Example:**

```bash
SUPABASE_DB_HOST=db.example.com
SUPABASE_DB_USER=myuser
SUPABASE_DB_PASSWORD=mypassword
SUPABASE_DB_NAME=mydb
SUPABASE_DB_PORT=5432
```

**Note:** If `SUPABASE_DB_PORT` is set to anything other than `5432`, the application will exit with an error. Only direct connections are supported.

### Connection Configuration

The backend uses the following settings for direct connections:

- **Port:** 5432 (required, validated at startup)
- **Protocol:** Standard PostgreSQL protocol with prepared statements enabled
- **SSL Mode:** `require` (default, can be overridden in connection string)
- **IPv6 Required:** Supabase direct connections (port 5432) require IPv6 connectivity
- **Connection Pool:**
  - Max idle connections: 5
  - Max open connections: 20
  - Connection max lifetime: 1 hour

### IPv6 Connectivity Requirement

**Important:** Supabase direct database connections (port 5432) require IPv6 connectivity. If you encounter a "network is unreachable" error, verify:

1. **Check IPv6 is enabled:**

   ```bash
   ip -6 addr show
   ```

2. **Test IPv6 connectivity:**

   ```bash
   ping6 -c 3 ipv6.google.com
   ```

3. **Verify DNS resolution:**

   ```bash
   host db.pnmqjubeshefgzboerss.supabase.co
   ```

4. **Check network configuration:** Ensure your network and firewall allow IPv6 connections

If IPv6 is not available on your system or network, you may need to:

- Enable IPv6 in your system settings
- Configure your network to support IPv6
- Use a different network environment that supports IPv6

### Error Messages

If you see an error about port validation:

```
Error: Only direct connections are supported. Port must be 5432, got: 6543
```

This means your connection string or `SUPABASE_DB_PORT` is set to 6543. Change it to 5432 to use a direct connection.

### Environment File

The application automatically loads environment variables from a `.env` file in the project root (if present). This is useful for local development. You can create a `.env` file with your database configuration:

```bash
# .env file example
DATABASE_URL=postgres://user:password@host:5432/dbname?sslmode=require
```

Or using individual components:

```bash
# .env file example
SUPABASE_DB_HOST=db.example.com
SUPABASE_DB_USER=myuser
SUPABASE_DB_PASSWORD=mypassword
SUPABASE_DB_NAME=mydb
```

**Important:** Environment variables from the system (e.g., provided by deployment platforms) always take precedence over values in `.env` files. This ensures that production configurations are never overridden by local development files.

## Deployment Platforms

### Coolify Support

This application is fully compatible with [Coolify](https://coolify.io) and other container-based deployment platforms. The application automatically reads environment variables provided by Coolify at runtime.

**How it works:**

1. **Local Development:** The app attempts to load a `.env` file for convenience during local development
2. **Production (Coolify):** Environment variables are provided directly by Coolify and automatically used by the application
3. **Precedence:** System environment variables (from Coolify) always override `.env` file values

**Setting Environment Variables in Coolify:**

1. Navigate to your application in Coolify
2. Go to the "Environment Variables" section
3. Add the required environment variables (see "Required Environment Variables" above)
4. The application will automatically use these variables on the next deployment

**Required Environment Variables for Coolify:**

Set these in Coolify's environment variables section:

- `DATABASE_URL` or `SUPABASE_DB_URL` (full connection string), OR
- `SUPABASE_DB_HOST`, `SUPABASE_DB_USER`, `SUPABASE_DB_PASSWORD` (individual components)
- `PORT` (optional, defaults to 8080)
- `ACCEPTED_ORIGINS` (comma-separated list of allowed CORS origins)
- `BACKEND_PASSWORD` (for API authentication)
- `CONTACT_RECIPIENT_EMAIL` (for contact form and newsletter endpoints)
- `RESEND_API_KEY` (for email sending via Resend API)
- `RESEND_FROM_EMAIL` (for email sending via Resend API)
- Any other service-specific variables (e.g., `MEDIUM_INTEGRATION_TOKEN`, etc.)

**Additional Environment Variables:**

- `GENERATE_MODELS` - Set to `true` to generate database models (development only)
- `GENERATE_COLUMN_REPORT` - Set to `true` to generate column mismatch report (development only)

The application will automatically detect and use environment variables provided by Coolify without requiring any `.env` file.

## Running the Application

### Prerequisites

- Go 1.21 or later
- PostgreSQL database (Supabase or other)
- Environment variables configured (see above)

### Build and Run

```bash
go build -o backend
./backend
```

Or run directly:

```bash
go run main.go
```

### Model Generation

To generate database models:

```bash
GENERATE_MODELS=true go run main.go
```

To generate a column mismatch report:

```bash
GENERATE_COLUMN_REPORT=true go run main.go
```

## API Documentation

API documentation is available via Swagger at:

- Swagger UI: `http://localhost:8080/swagger/index.html`
- Swagger JSON: `http://localhost:8080/swagger/doc.json`

## Root Endpoint (Message of the Day)

The backend provides a root endpoint (`/`) that displays a "Message of the Day" with a fun, personalized touch. The response adapts based on the client:

- **Endpoint:** `GET /`
- **CORS:** Accessible from any origin (no authentication required)
- **Response Format:** 
  - **curl/command-line tools:** Plain text with a random status message
  - **Browsers:** HTML terminal-style interface with dark mode styling (Catppuccin colors)

**Custom Headers:**
- `X-Powered-By`: "Arch Linux (btw)"
- `X-Quantum-State`: "Superposition"

**Example Usage:**

```bash
# Plain text response for curl
curl http://localhost:8080/

# Example output:
# SYSTEM STATUS: ONLINE.
# WARNING: Cat detected in server room.
```

When accessed from a browser, the endpoint displays a styled terminal window with the status message and a blinking cursor.

**Available Messages:**
The endpoint cycles through various personalized messages including:
- System status updates
- Quantum physics references
- Arch Linux/Hyprland configuration notes
- Puzzle game references

## Healthcheck Endpoint

The backend provides a healthcheck endpoint that can be accessed from any origin:

- **Endpoint:** `GET /healthcheck`
- **CORS:** Accessible from any origin (no authentication required)
- **Response:** JSON object containing:
  - `current_time`: Current server date and time (RFC3339 format)
  - `startup_time`: Server startup time, representing when this version was deployed (RFC3339 format)
  - `uptime_seconds`: Server uptime in seconds

**Example Response:**

```json
{
  "current_time": "2024-01-15T10:30:45Z",
  "startup_time": "2024-01-15T10:00:00Z",
  "uptime_seconds": 1845
}
```

**Usage:**

```bash
curl http://localhost:8080/healthcheck
```

This endpoint is useful for monitoring server status and deployment verification.

## Contact Endpoints

The backend provides public endpoints for contact form submissions and newsletter subscriptions. These endpoints do not require authentication and are accessible from any origin (subject to CORS configuration).

### Submit Contact Form

- **Endpoint:** `POST /api/contact`
- **CORS:** Accessible from any origin (no authentication required)
- **Request Body:**
  ```json
  {
    "name": "John Doe",
    "email": "[email protected]",
    "subject": "Optional subject line",
    "message": "Your message here (minimum 10 characters)"
  }
  ```
- **Response:** JSON object containing:
  ```json
  {
    "success": true,
    "message": "Thank you for your message! I'll get back to you soon."
  }
  ```
- **Validation:**
  - `name`: Required, must not be empty
  - `email`: Required, must be a valid email address (contains "@")
  - `subject`: Optional
  - `message`: Required, must be at least 10 characters

**Example Usage:**
```bash
curl -X POST http://localhost:8080/api/contact \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "[email protected]",
    "subject": "Hello",
    "message": "This is a test message from the contact form."
  }'
```

**Error Responses:**
- `400 Bad Request`: Invalid or missing required fields
- `500 Internal Server Error`: Email service configuration error or email sending failure

### Subscribe to Newsletter

- **Endpoint:** `POST /api/newsletter/subscribe`
- **CORS:** Accessible from any origin (no authentication required)
- **Request Body:**
  ```json
  {
    "email": "[email protected]"
  }
  ```
- **Response:** JSON object containing:
  ```json
  {
    "success": true,
    "message": "Thank you for subscribing to the newsletter!"
  }
  ```
- **Validation:**
  - `email`: Required, must be a valid email address (contains "@")

**Example Usage:**
```bash
curl -X POST http://localhost:8080/api/newsletter/subscribe \
  -H "Content-Type: application/json" \
  -d '{
    "email": "[email protected]"
  }'
```

**Error Responses:**
- `400 Bad Request`: Invalid or missing email address
- `500 Internal Server Error`: Email service configuration error or email sending failure

### Contact Endpoint Configuration

The contact endpoints require the following environment variables:

- **`CONTACT_RECIPIENT_EMAIL`** (required): The email address where contact form submissions and newsletter subscriptions will be sent. This should be set to the site owner's email address.
- **`RESEND_API_KEY`** (required): Your Resend API key for sending emails. See the [Resend documentation](https://resend.com/docs) for more information.
- **`RESEND_FROM_EMAIL`** (required): The sender email address in the format "Your Name <[email protected]>". This is used as a fallback for `CONTACT_RECIPIENT_EMAIL` if not set (for development purposes).

**Note:** If `CONTACT_RECIPIENT_EMAIL` is not set, the endpoints will return a 500 error. The email service uses the Resend API to send formatted HTML emails with the contact form details or newsletter subscription information.

## Project Structure

```
backend/
├── api/           # HTTP handlers, routes, middleware
├── config/        # Configuration management
├── database/      # Database repositories and connection management
├── docs/          # Swagger/OpenAPI documentation
├── errs/          # Error definitions
├── models/        # Data models and database schemas
├── services/      # Business logic and external service integrations
└── main.go        # Application entry point
```
