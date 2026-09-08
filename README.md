# Eventura — Event Management Platform (MVP)

A full-stack event ticketing platform: organizers create and promote events, attendees browse, search, and buy tickets. Built as a monorepo with **Next.js** (frontend), **Go** (backend API), and **MySQL** (database), with payments via **Midtrans Snap** (sandbox).

## Tech stack

| Layer | Choice |
|---|---|
| Frontend | Next.js 16 (App Router, TypeScript), Tailwind CSS v4, React Hook Form + Zod, Recharts, `jose` |
| Backend | Go, Gin, GORM (MySQL driver), JWT (`golang-jwt/jwt`), bcrypt |
| Database | MySQL 8 (schema managed via GORM `AutoMigrate` on boot) |
| Payments | Midtrans Snap (sandbox) |
| Tests | Go `testing` + an in-memory pure-Go SQLite DB (backend), Jest + React Testing Library (frontend) |

## Project layout

```
event-management/
├── backend/            Go API (cmd/api, internal/{config,models,handlers,services,middleware,...})
├── frontend/            Next.js app (src/app, src/components, src/lib, src/context)
├── docker-compose.yml   MySQL + backend + frontend, one command
└── .env.example         Root env vars consumed by docker-compose
```

## Quick start (Docker Compose — recommended)

1. Copy the root env file and fill in your Midtrans sandbox keys (get them free at https://dashboard.sandbox.midtrans.com/settings/config_info — register an account, no real business required):

   ```bash
   cp .env.example .env
   # edit .env: set JWT_SECRET to any long random string, and MIDTRANS_SERVER_KEY / MIDTRANS_CLIENT_KEY
   ```

2. Build and start everything:

   ```bash
   docker compose up --build
   ```

3. Open:
   - Frontend: http://localhost:3000
   - Backend API: http://localhost:8080/api/v1 (health check at http://localhost:8080/health)

   MySQL data and uploaded event banners persist in named Docker volumes (`mysql_data`, `uploads_data`) across restarts.

Paid-event checkout requires valid Midtrans sandbox keys; free events and every other feature (auth, search/filter/pagination, vouchers, referrals, points, reviews, the organizer dashboard) work without them.

## Manual setup (without Docker)

**Prerequisites:** Go 1.26+, Node 20+, a running MySQL 8 instance.

### Backend

```bash
cd backend
cp .env.example .env         # edit DB_* to point at your local MySQL, and set JWT_SECRET
go mod download
go run ./cmd/api              # http://localhost:8080
```

The API creates the database schema automatically on first boot (`AutoMigrate`) and seeds a default list of categories.

### Frontend

```bash
cd frontend
cp .env.local.example .env.local   # JWT_SECRET must match the backend's exactly
npm install
npm run dev                         # http://localhost:3000
```

## Running tests

```bash
# Backend — unit tests for auth, events, vouchers, referral/points, checkout
# (including the SQL-transaction rollback and expiry-sweep paths), and reviews.
cd backend
go test ./...

# Frontend — the debounce hook, checkout pricing math, and key UI flows
# (confirm dialogs, pagination, star rating, login form validation).
cd frontend
npm test
```

## Feature checklist

- **Event discovery** — landing page lists published events; search bar (debounced, 400ms), category/city/price filters, sort, and pagination; empty states for no-results. `frontend/src/app/page.tsx`
- **Event details & booking** — full event page with ticket type selection, live price preview (subtotal → coupon → points), and Midtrans Snap checkout for paid events (instant confirmation for free/fully-discounted orders). `frontend/src/app/events/[slug]`, `frontend/src/components/BookingPanel.tsx`
- **Vouchers** — organizers create quota-limited, date-windowed discount codes per event. `backend/internal/handlers/voucher_handler.go`
- **Reviews & ratings** — attendees can rate/comment once an event has ended and only for orders they completed. `backend/internal/handlers/review_handler.go`
- **Auth & roles** — customer vs organizer signup, JWT auth, and route protection via a Next.js proxy (`middleware`) that verifies the JWT and redirects by role. `frontend/src/proxy.ts`
- **Referral, points & coupons** — referral code generated per user; referring a signup earns the referrer 10,000 points (expiring in 3 months) and grants the new user a 10%-off coupon (also 3 months). Points redeem 1:1 against IDR at checkout. `backend/internal/services/points_service.go`
- **Organizer dashboard** — event/attendee/transaction management and revenue charts bucketed per year (by month), per month (by day), or per day (by hour). `frontend/src/app/organizer/dashboard`, `backend/internal/handlers/dashboard_handler.go`
- **SQL transactions** — every multi-row write (registration + referral bonus, checkout, cancel, Midtrans webhook, expiry sweep) runs inside a single DB transaction with row locking (`SELECT ... FOR UPDATE`) so seat counts, voucher/coupon usage, and point ledgers never drift under concurrent requests. `backend/internal/services/transaction_service.go`
- **Confirmation dialogs** — deleting an event/voucher and cancelling an order all go through a shared `useConfirmDialog` popup before the request fires. `frontend/src/components/ui/ConfirmDialog.tsx`

## Notable design decisions & trade-offs

- **Auth token storage.** The JWT is kept in a regular (non-`httpOnly`) cookie so both the browser's `axios` client (`Authorization: Bearer …`) and the Next.js proxy (signature-verified via `jose`) can read it without a server-side proxy for every API call. This keeps the app simple for an MVP; a production build should move to an `httpOnly` cookie plus a server-side API proxy to fully mitigate XSS token theft.
- **Point ledger.** Points are an append-only ledger (`earn`/`redeem` rows); balance is the sum of unexpired `earn` rows minus all `redeem` rows. This is simpler than tracking per-batch remaining balances and FIFO expiry, at the cost of not expiring "used" points precisely in creation order — acceptable for the MVP's stated rules.
- **Payment window.** Orders reserve seats/discounts immediately at checkout and hold them for `PAYMENT_DEADLINE_MINUTES` (default 120). A background sweeper (`internal/jobs/expiry.go`) releases stale holds every 5 minutes; Midtrans's webhook finalizes success/failure as soon as it fires.
- **Uploads.** Event banners are stored on local disk under `backend/uploads/` (served at `/uploads/...`, persisted via a Docker volume in Compose). Fine for a single-instance MVP; swap for S3/GCS if you scale to multiple backend instances.

## API overview

All endpoints are under `/api/v1`. Highlights:

- `POST /auth/register`, `POST /auth/login`, `GET /auth/me`, `GET /referral/me`
- `GET /events`, `GET /events/:slug`, `GET /events/:slug/reviews`, `GET /categories`
- `POST /transactions`, `GET /transactions/me`, `POST /transactions/:id/cancel`, `POST /midtrans/notification`
- `POST /reviews`, `GET /reviews/reviewable`
- `POST|PUT|DELETE /organizer/events`, `GET /organizer/events/:id/attendees`
- `POST|GET|DELETE /organizer/events/:id/vouchers`
- `GET /organizer/transactions`, `GET /organizer/dashboard/stats?range=year|month|day`
- `POST /organizer/uploads/banners`, `POST /uploads/avatars`

See `backend/internal/server/router.go` for the full route table.
