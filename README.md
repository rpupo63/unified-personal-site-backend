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

The application automatically loads environment variables from a `.env` file in the project root (if present). You can create a `.env` file with your database configuration:

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
