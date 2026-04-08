# be-zor

Backend service for the finance dashboard described in [`TASK_CONTEXT.md`](../TASK_CONTEXT.md).

This service is responsible for authentication, session validation, role-based access control, transaction storage, dashboard summaries, analyst reporting, and admin management. The backend owns the business rules; the frontend only consumes the APIs.

## Project Context Coverage

The backend matches the task brief directly:

- **User and role management**
  Supports `viewer`, `analyst`, and `admin` roles plus `active` and `inactive` account states.
- **Financial records management**
  Supports create, read, update, delete, and filter operations for transactions.
- **Dashboard summary APIs**
  Calculates totals, category summaries, recent activity, monthly trends, and weekly trends on the server.
- **Access control logic**
  Enforced in Fiber middleware, not only in the UI.
- **Validation and error handling**
  Request payloads, filters, roles, status, dates, and credentials are validated in the backend.
- **Persistence**
  Uses CockroachDB with Bun ORM and SQL migrations embedded into the Go binary.

## Stack Used

This backend currently uses:

- **Go 1.25**
- **Fiber v2** for HTTP routing and middleware
- **Bun ORM** for database access
- **Bun migrate** for schema migrations
- **CockroachDB** as the relational database
- **bcrypt** for local password hashing
- **Google ID token verification** using Google public certificates
- **Docker Compose** for local database setup

Project-adjacent integration used by this repo:

- **Next.js 16** frontend in [`../zorvyn`](../zorvyn)
- **NextAuth** in the frontend to bridge Google login and manual local login into backend sessions

## What The Backend Owns

This service is not just CRUD. It is the policy layer for the project. It owns:

- who can access which route
- whether a session is valid
- whether a user account is active
- which transactions each role is allowed to see
- how summary data is calculated
- how admin actions are validated

## Runtime Architecture

### Entry point

- [`main.go`](./main.go)
  Loads config, opens the database, optionally runs migrations, registers middleware, wires handlers, and starts Fiber.

### Config

- [`internal/config/config.go`](./internal/config/config.go)
  Loads environment variables from process env and `.env` files.

### Database

- [`internal/database/db.go`](./internal/database/db.go)
  Opens the Bun connection and creates the target database if needed.
- [`internal/database/migrations.go`](./internal/database/migrations.go)
  Embeds and runs SQL migrations.
- [`cmd/migrate/main.go`](./cmd/migrate/main.go)
  CLI for `up`, `down`, and `status`.

### Models

- [`internal/models/auth.go`](./internal/models/auth.go)
  User, session, Google auth, and local auth models.
- [`internal/models/transaction.go`](./internal/models/transaction.go)
  Transaction models, filters, summaries, and admin/analyst DTOs.
- [`internal/models/admin.go`](./internal/models/admin.go)
  Admin user request and response contracts.

### Middleware

- [`internal/middleware/auth.go`](./internal/middleware/auth.go)
  Bearer token parsing, session validation, account status enforcement, and role checks.

### Handlers

- [`internal/handlers/auth.go`](./internal/handlers/auth.go)
  Google auth and `/api/me`.
- [`internal/handlers/local_auth.go`](./internal/handlers/local_auth.go)
  Local signup and signin.
- [`internal/handlers/dashboard.go`](./internal/handlers/dashboard.go)
  Dashboard summary route.
- [`internal/handlers/transactions.go`](./internal/handlers/transactions.go)
  Current-user transaction listing.
- [`internal/handlers/analyst.go`](./internal/handlers/analyst.go)
  Analyst overview route.
- [`internal/handlers/admin.go`](./internal/handlers/admin.go)
  Admin user and transaction management.
- [`internal/handlers/transaction_filters.go`](./internal/handlers/transaction_filters.go)
  Shared query filter parsing.

### Store layer

- [`internal/store/bun.go`](./internal/store/bun.go)
  Core auth, session, and per-user transaction operations.
- [`internal/store/users.go`](./internal/store/users.go)
  Admin user management queries.
- [`internal/store/transactions.go`](./internal/store/transactions.go)
  Cross-user transaction queries and transaction mutations.

### Summary layer

- [`internal/summary/calculator.go`](./internal/summary/calculator.go)
  Summary math and trend generation.
- [`internal/summary/access.go`](./internal/summary/access.go)
  Organization-level summaries and analyst anonymization output.

### External auth verification

- [`internal/googleauth/verifier.go`](./internal/googleauth/verifier.go)
  Verifies Google JWTs against Google certificate keys and validates issuer, audience, and expiry.

## Database Design

The schema is defined in [`db/schema.sql`](./db/schema.sql).

### `users`

Stores application users for both Google and local auth.

Important fields:

- `provider`
- `role`
- `status`
- `password_hash`
- `google_subject`
- `email`
- `name`
- `last_login_at`

Design choices:

- `role` is stored directly on the user as `viewer`, `analyst`, or `admin`
- `status` is stored directly on the user as `active` or `inactive`
- local accounts use `password_hash`
- Google users use Google identity metadata and `google_subject`

### `sessions`

Stores backend-managed sessions independently of the frontend session.

Important fields:

- `id`
- `user_id`
- `token`
- `provider`
- `expires_at`
- `last_used_at`
- `user_agent`
- `ip_address`

Design choices:

- protected routes require both the token and the session ID
- session expiry is enforced in the backend
- `last_used_at` is refreshed on validated requests
- sessions cascade on user deletion

### `transactions`

Stores finance records owned by users.

Important fields:

- `user_id`
- `amount`
- `type`
- `category`
- `transaction_date`
- `description`

Design choices:

- `type` is constrained to `income` or `expense`
- `amount` must be non-negative at the database level
- indexes support common filtering and sorting paths

## Migration History

Current migrations:

- `202604040001_initial_auth`
- `202604040002_add_transactions`
- `202604050001_add_roles`
- `202604050002_simplify_roles_to_user_field`
- `202604060001_add_local_password_auth`
- `202604060002_add_user_status`

This history reflects the evolution from an early roles table design to the simpler role-on-user model that better matches the project scope.

## Role Model

### `viewer`

- Can access their own dashboard summary
- Can access their own transaction records
- Cannot access analyst or admin routes

### `analyst`

- Can access organization-level overview data
- Can view cross-user financial groupings through analyst routes
- Cannot create, update, or delete transactions
- Cannot manage users

### `admin`

- Can manage users
- Can manage all transactions
- Can access analyst routes as well

## Account Status Model

User status is separate from session validity.

- `active`: allowed to use the system
- `inactive`: blocked even if an old session token still exists

This matters because an admin can deactivate a user and the backend will reject future protected requests for that account.

## Authentication Design

The project supports two auth flows:

- Google sign-in/sign-up
- Local email/password sign-up and sign-in

Both end in the same backend session model.

### Google auth flow

1. The frontend sends a Google ID token to the backend.
2. The backend verifies the JWT signature and claims.
3. The backend creates or updates the user.
4. The backend creates a session.
5. The backend returns `sessionToken`, `session`, and `user`.

Verified conditions:

- audience must match `GOOGLE_CLIENT_ID`
- issuer must be a valid Google issuer
- token must not be expired
- token issue time must be valid

### Local auth flow

1. The client sends a signup or signin payload.
2. The handler validates required fields.
3. Passwords are hashed and checked with bcrypt.
4. The backend creates a session on success.

Validation currently includes:

- required name and email for signup
- required password fields
- password minimum length of 8
- signup password confirmation
- email normalization and validation

### Session model

Protected backend routes require:

- `Authorization: Bearer <session-token>`
- `X-Session-ID: <session-id>`

The backend validates both values together. If the session is missing, expired, or mismatched, the request is rejected.

## Frontend Integration In This Repo

The backend is designed to work with the frontend in [`../zorvyn`](../zorvyn).

Important repo-level integration details:

- the frontend uses **NextAuth**
- Google login is synced into the backend by posting the Google ID token to `/api/auth/google/signup`
- local auth is bridged into NextAuth using backend-issued session credentials
- server-side frontend requests include both `Authorization` and `X-Session-ID`
- the backend can load `GOOGLE_CLIENT_ID` from `../zorvyn/.env` if it is not present in `be-zor/.env`

## Route Map

All routes are registered in [`main.go`](./main.go).

### Health

| Method | Route | Purpose |
| --- | --- | --- |
| `GET` | `/` | simple service status response |

### Public auth routes

| Method | Route | Purpose |
| --- | --- | --- |
| `POST` | `/api/auth/google/signup` | verify Google token, upsert user, create backend session |
| `POST` | `/api/auth/google/signin` | currently uses the same handler as Google signup |
| `POST` | `/api/auth/local/signup` | create local user and backend session |
| `POST` | `/api/auth/local/signin` | authenticate local user and create backend session |

### Protected routes for any authenticated active user

| Method | Route | Purpose |
| --- | --- | --- |
| `GET` | `/api/me` | return authenticated user and session |
| `GET` | `/api/dashboard/summary` | return user-level dashboard summary |
| `GET` | `/api/transactions` | return only the authenticated user’s transactions |

### Analyst routes

Protected by `RequireAuth` and `RequireRole(analyst, admin)`.

| Method | Route | Purpose |
| --- | --- | --- |
| `GET` | `/api/analyst/overview` | organization summary plus anonymized grouped records |

### Admin routes

Protected by `RequireAuth` and `RequireRole(admin)`.

#### User management

| Method | Route | Purpose |
| --- | --- | --- |
| `GET` | `/api/admin/users` | list users with transaction counts |
| `POST` | `/api/admin/users` | create a local user with role and status |
| `PATCH` | `/api/admin/users/:userID` | update name, email, password, role, and status |
| `DELETE` | `/api/admin/users/:userID` | delete a user and cascade related data |

#### Transaction management

| Method | Route | Purpose |
| --- | --- | --- |
| `GET` | `/api/admin/transactions` | list all transactions with filters |
| `POST` | `/api/admin/transactions` | create a transaction for any user |
| `PATCH` | `/api/admin/transactions/:transactionID` | update any transaction |
| `DELETE` | `/api/admin/transactions/:transactionID` | delete any transaction |

## Filtering

Filtering is implemented in the backend and reused across:

- `/api/transactions`
- `/api/analyst/overview`
- `/api/admin/transactions`

Supported query parameters:

- `dateFrom`
- `dateTo`
- `category`
- `type`

Rules:

- dates must use `YYYY-MM-DD`
- `dateFrom` cannot be later than `dateTo`
- `type` must be `income` or `expense`
- category filtering is case-insensitive

## Summary Logic

The summary layer returns dashboard-friendly output instead of forcing the frontend to aggregate raw records.

Current summary output includes:

- total income
- total expenses
- net balance
- category totals
- recent activity
- monthly trends
- weekly trends

Important behavior:

- recent activity is limited to 5 records
- monthly trends are grouped by `YYYY-MM`
- weekly trends are grouped by ISO week
- analyst overview uses organization-wide data

## Analyst Anonymization

The analyst response intentionally avoids direct user identity.

Current behavior:

- transactions are grouped internally by real `user_id`
- the response exposes generated labels such as `Account 1`, `Account 2`
- analyst payloads do not return real user name or email

This keeps analyst access useful for reporting while reducing identity exposure.

## Validation And Guard Rails

Examples of backend protections already implemented:

- invalid request bodies return `400`
- invalid credentials return `401`
- inactive users return `403`
- missing role permissions return `403`
- duplicate emails return `409`
- missing users or transactions return `404`

Admin-specific guard rails:

- admin cannot delete their own account
- admin cannot remove their own admin role
- admin cannot deactivate their own account
- password updates are only allowed for local users

## Environment Variables

Configuration is loaded by [`internal/config/config.go`](./internal/config/config.go).

### Backend variables

| Variable | Required | Default | Purpose |
| --- | --- | --- | --- |
| `DATABASE_URL` | yes | none | CockroachDB/Postgres-style DSN |
| `PORT` | no | `8080` | backend server port |
| `FRONTEND_ORIGIN` | no | `http://localhost:3000` | allowed CORS origin |
| `SHOULD_MIGRATE` | no | `false` | run migrations automatically at startup |
| `SESSION_TTL_HOURS` | no | `24` | session lifetime in hours |
| `GOOGLE_CLIENT_ID` | yes for Google auth | none | expected Google audience for ID token verification |

### Current local backend `.env`

The repo currently uses a backend `.env` with:

```env
DATABASE_URL=postgresql://root@localhost:26257/finance_db?sslmode=disable
PORT=8080
SHOULD_MIGRATE=true
```

Recommended expanded local example:

```env
DATABASE_URL=postgresql://root@localhost:26257/finance_db?sslmode=disable
PORT=8080
FRONTEND_ORIGIN=http://localhost:3000
SHOULD_MIGRATE=true
SESSION_TTL_HOURS=24
GOOGLE_CLIENT_ID=your-google-client-id.apps.googleusercontent.com
```

Note: if `GOOGLE_CLIENT_ID` is missing in `be-zor/.env`, the backend also tries to read it from `../zorvyn/.env`.

## Local Development

### 1. Start CockroachDB

From the repo root:

```bash
docker compose up -d
```

This uses the root-level [`../docker-compose.yml`](../docker-compose.yml) to start a single-node CockroachDB instance on:

- `26257` for SQL
- `8081` for the Cockroach admin UI

### 2. Start the backend

From [`be-zor`](./):

```bash
go run .
```

If `SHOULD_MIGRATE=true`, migrations run automatically before the server starts.

### 3. Migration commands

```bash
go run ./cmd/migrate up
go run ./cmd/migrate down
go run ./cmd/migrate status
```

## Example Request Shapes

### Local signup

```json
{
  "name": "Local User",
  "email": "local@example.com",
  "password": "manualpass123",
  "confirmPassword": "manualpass123"
}
```

### Local signin

```json
{
  "email": "local@example.com",
  "password": "manualpass123"
}
```

### Google signup/signin

```json
{
  "credential": "<google-id-token>"
}
```

### Authenticated request headers

```http
Authorization: Bearer <session-token>
X-Session-ID: <session-id>
```

### Transaction filters

```text
/api/transactions?dateFrom=2026-04-01&dateTo=2026-04-30&category=salary&type=income
```

## Current Scope Notes

Implemented well for the project brief:

- backend-managed auth and sessions
- role-based access control
- admin and analyst separation
- backend summaries
- backend filtering
- migration-based schema evolution

Not currently implemented:

- automated tests
- pagination
- rate limiting
- OpenAPI/Swagger docs
- refresh token flow

## Why This Design Fits The Task

This backend fits the project context because it prioritizes business rules over superficial endpoint count. The important decisions all live on the server:

- access is enforced in middleware
- summaries are computed in backend code
- account status can override session possession
- analyst access is useful but intentionally anonymized
- admin actions are protected by backend guard rails

The result is a backend that is aligned with the finance dashboard brief and with the actual frontend integration already present in this repository.
