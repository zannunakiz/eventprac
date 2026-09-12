<p align="center">
  <img src="https://img.shields.io/badge/Go-1.26.0-00ADD8?logo=go&logoColor=white" alt="Go 1.26">
  <img src="https://img.shields.io/badge/Web%20Framework-Gin%201.12-00A6D6" alt="Gin">
  <img src="https://img.shields.io/badge/ORM-GORM%201.31-3E8EED" alt="GORM">
  <img src="https://img.shields.io/badge/Database-PostgreSQL-336791?logo=postgresql&logoColor=white" alt="PostgreSQL">
  <img src="https://img.shields.io/badge/Hosted%20on-NeonDB-00E1B8" alt="NeonDB">
  <img src="https://img.shields.io/badge/Auth-JWT%20%2B%20bcrypt-8A2BE2" alt="JWT + bcrypt">
  <img src="https://img.shields.io/badge/Media%20Storage-ImageKit-EA4C89" alt="ImageKit">
  <img src="https://img.shields.io/badge/API%20Testing-Postman-FF6C37" alt="Postman">
</p>

<h1 align="center">🎟️ Eventprac — Event Booking API</h1>

<p align="center">
  A practice-grade <b>event booking REST API</b> — users register, log in, create events with cover images,
  and book other people's events.<br/>
  <i>Go (lang) · Gin · GORM · PostgreSQL (Neon) · JWT · ImageKit</i>
</p>

---

## Table of Contents

- [About](#about)
- [Features](#features)
- [Data Model](#data-model)
- [Getting Started](#getting-started)
- [Environment Variables](#environment-variables)
- [Database & Auto-Migration](#database--auto-migration)
- [API Reference](#api-reference)
- [Authentication](#authentication)
- [Postman Collection](#postman-collection)
- [Project Structure](#project-structure)
- [License & Disclaimer](#license--disclaimer)

---

## About

A REST API for **creating, discovering, booking, and managing events** — built as a hands-on practice project to sharpen skills with the Go (lang) ecosystem: **Gin** for HTTP, **GORM** for ORM/auto-migration, **JWT + bcrypt** for auth, **Neon** for serverless PostgreSQL, and **ImageKit** for image CDN.

```
  Client ⇄ Gin (main.go) ⇄ JWT Middleware ⇄ Controllers ⇄ GORM ⇄ Neon / PostgreSQL
                                          └────────────────────────⇄ ImageKit (images)
```

**Highlights**
- JWT (HS256) auth with bcrypt-hashed passwords — tokens valid for **7 days**
- Event CRUD with optional cover-image uploads to ImageKit
- Search (`ILIKE` on name/description) + pagination out of the box
- Booking system with unique human-readable booking codes
- Auto-migration on boot — no manual SQL required
- Health endpoint that pings the database & ImageKit
- Postman collection + automated test suite bundled

---

## Features

| Area | What it does |
|---|---|
| 🔐 Auth | `register` / `login` / `me` — bcrypt hashing, 7-day JWT (HS256) |
| 📅 Events | Full CRUD with image upload to ImageKit, `?search=` filtering, `?page=`/`?limit=` pagination |
| 🗂️ Ownership | Only the event creator can update or delete their event — enforced server-side |
| 🎫 Bookings | Book / list / cancel; duplicate booking (same user + event) rejected; auto `BK-…` codes |
| 🖼️ Media hygiene | Images are deleted from ImageKit when validation/DB steps fail, or when replaced/deleted |
| 🩹 Observability | `GET /health` reports `up` / `down` per service (server, database, imagekit) |
| 🧹 Dev utility | `POST /clearall` wipes all tables + ImageKit files — gated behind a secret code |

---

## Data Model

| Model | Fields |
|---|---|
| **User** | `id` · `name` · `email` *(unique)* · `password` *(bcrypt hash, never serialized)* |
| **Event** | `id` · `name` · `description` · `location` · `datetime` *(RFC 3339)* · `image` · `imageId` · `userId` *(FK)* |
| **Booking** | `id` · `bookingCode` *(auto-generated `BK-…`)* · `phone` · `userId` *(FK)* · `eventId` *(FK)* |

**Relationships**

```
User   1 ── N Event      (a user can create many events)
User   1 ── N Booking    (a user can make many bookings)
Event  1 ── N Booking    (an event can have many bookings)
```

---

## Getting Started

### Prerequisites

- **Go (lang) v1.26+** — https://golang.org/dl
- **Neon** Postgres database — https://neon.tech (any Postgres works too)
- **ImageKit** account (free tier) — https://imagekit.io

### 1. Clone

```bash
git clone https://github.com/zannunakiz/eventprac.git
cd eventprac
```

### 2. Configure environment

```bash
cp env.example .env
```

Fill in every variable — see [Environment Variables](#environment-variables).

### 3. Run

```bash
go run .
```

That's it. On startup the server will:

1. load `.env`
2. connect to Neon/Postgres
3. **auto-migrate** the schema (`users`, `events`, `bookings` — check your Neon dashboard ✓)
4. listen on **`http://localhost:8080`**

---

## Environment Variables

| Variable | Required | Description |
|---|---|---|
| `DATABASE_URI` | ✅ | Postgres connection string (Neon pooler string works out of the box) |
| `JWT_SECRET` | ✅ | Random string used to sign/verify JWT tokens |
| `IMAGEKIT_PUBLIC_KEY` | ✅ | ImageKit dashboard → Developer Options |
| `IMAGEKIT_PRIVATE_KEY` | ✅ | ImageKit dashboard → Developer Options |
| `IMAGEKIT_URL_ENDPOINT` | ✅ | e.g. `https://ik.imagekit.io/<your-id>` |
| `CLEAR_ALL_CODE` | ✅ | Secret code required to call the experimental `POST /clearall` |

**Example**

```bash
DATABASE_URI=postgresql://user:pass@ep-xxxx.aws.neon.tech/neondb?sslmode=require
JWT_SECRET=change-me-to-a-long-random-string
IMAGEKIT_PUBLIC_KEY=public_xxxx
IMAGEKIT_PRIVATE_KEY=private_xxxx
IMAGEKIT_URL_ENDPOINT=https://ik.imagekit.io/xxxx
CLEAR_ALL_CODE=my-secret-clear-code
```

> ⚠️ `.env` is git-ignored. Never commit real credentials.

---

## Database & Auto-Migration

- Connections use a standard Postgres DSN; `sslmode=require` is recommended for Neon.
- **GORM's `AutoMigrate`** (`config/db.go`) creates/updates tables on every boot from the models in `models/*.go` — no manual SQL or migration files needed.
- To reset everything during development, call `POST /clearall` (it also purges ImageKit files).

---

## API Reference

Base URL: `http://localhost:8080` — errors return `{ "error": "…" }`. 🔒 = JWT `Bearer` token required.

| Method | Endpoint | Auth | Description |
|---|---|---|---|
| GET | `/health` | — | Health of server, database & ImageKit (`200`/`503`) |
| POST | `/clearall` | — | Wipe all data — body `{ "clearall-code": "…" }` (DEV ONLY) |
| POST | `/api/user/register` | — | Create account — `{ name, email, password }` |
| POST | `/api/user/login` | — | Returns `token` (7-day JWT) + user |
| GET | `/api/user/me` | 🔒 | Current user profile |
| GET | `/api/events` | — | List events — `?search=&page=1&limit=5` |
| GET | `/api/events/:id` | — | Single event (incl. creator + bookings) |
| GET | `/api/events/user` | 🔒 | My events |
| POST | `/api/events` | 🔒 | Create event — `multipart`, image optional |
| PATCH | `/api/events/:id` | 🔒 | Update event (owner) — new image replaces old |
| DELETE | `/api/events/:id` | 🔒 | Delete event (owner) |
| POST | `/api/booking` | 🔒 | Book event — `{ eventId, phone }` |
| GET | `/api/booking/user` | 🔒 | My bookings |
| DELETE | `/api/booking/:id` | 🔒 | Cancel booking (owner) |

**Example (login → create event):**

```bash
TOKEN=<jwt from POST /api/user/login>

curl -X POST http://localhost:8080/api/user/login -H "Content-Type: application/json" -d '{"email":"alice@example.com","password":"secret123"}'

curl -X POST http://localhost:8080/api/events -H "Authorization: Bearer $TOKEN" -F "name=Jazz Night" -F "description=Live jazz" -F "datetime=2026-12-25T19:00:00+08:00"
```

---

## Authentication

- All protected routes require the header `Authorization: Bearer <token>`.
- Tokens are **JWT HS256** signed with `JWT_SECRET`, containing `sub` (user id) and `exp` (**now + 7 days**).
- Passwords are never stored in plain text — hashed with **bcrypt** (default cost).
- `middlewares/auth.middleware.go` validates the token and injects the resolved user id into the request context for controllers.

---

## Postman Collection

Two companion files are bundled:

### `postman-collection.json` — Manual API Explorer
Every endpoint, pre-wired with collection variables: `{{baseUrl}}` (default `http://localhost:8080`), `{{token}}`, `{{eventId}}`, `{{bookingId}}`, `{{clearAllCode}}`. Best for hand-clicking through the API and inspecting responses.

### `postman-test.json` — Full E2E Automated Test Suite (55 tests)
An ordered, top-to-bottom suite built for the **Collection Runner**. It doesn't only check happy paths — it deliberately **probes the flaws**: bad payloads, missing/malformed tokens, duplicate records, and cross-user access all must fail with the correct status code.

**Coverage — 8 sections, 55 tests:**

| Section | What it verifies |
|---|---|
| 🩹 01 Health (1) | `/health` → `200` + `healthy`, all services `up` |
| 🔐 02 Auth (13) | register `201`; rejects duplicate email, bad email, short/empty payload; login vs wrong password, unknown email, missing field → `401` |
| 👤 02b Me (2) | valid token returns user; missing / invalid / malformed `Authorization` → `401` |
| 📄 03 Event reads (4) | public list `200`; non-existent event `404`; creation & own-events are auth-gated |
| ✍️ 04 Owner CRUD (11) | multipart create, invalid `datetime` → `400`, fetch, search, pagination, update, delete flow |
| 🔑 05 Ownership (5) | a **second user** cannot update or delete another's event → `403` |
| 🎫 06 Bookings (13) | no-token `401`, missing fields `400`, unknown event `404`, **duplicate booking rejected**, list, booking-delete ownership |
| 🧹 07/08 Teardown (6) | full update, clean delete, confirm gone; final `/clearall` wipe + verify tables/users are empty |

**Security & validation scenarios it catches:**

- **Auth bypass** — every protected route is hit *without*, *with invalid*, and *with malformed* tokens (expect `401`)
- **Duplicate integrity** — re-registering an email and booking the same event twice both fail with `400`
- **Cross-user isolation** — updating/deleting someone else's event or booking → `403`
- **No secret leakage** — asserts `user.password` is never present in any response
- **Input validation** — short password, empty body, bad email, invalid RFC 3339 `datetime` → `400`
- **404 handling** for missing events/bookings

**Run it:**
1. Set the `clearAllCode` collection variable to your `CLEAR_ALL_CODE` from `.env` — otherwise the final wipe step fails.
2. Postman → **Import** → `postman-test.json`.
3. **Collection Runner** → run top-to-bottom. Note: Create/Update Event send **`multipart/form-data`, not JSON**.

---

## Project Structure

```
.
├── main.go                     # App entry — env, DB connect, route wiring
├── config/
│   └── db.go                   # Neon/Postgres connection + AutoMigrate
├── controllers/
│   ├── system.controller.go    # /health, /clearall
│   ├── user.controller.go      # register, login, me
│   ├── event.controller.go     # event CRUD + ImageKit uploads
│   └── booking.controller.go   # booking create / list / delete
├── middlewares/
│   └── auth.middleware.go      # JWT Bearer verification
├── models/
│   ├── user.go                 # User entity (unique email)
│   ├── event.go                # Event entity
│   └── booking.go              # Booking entity
├── env.example                 # Environment template
├── postman-collection.json     # Postman API collection
├── postman-test.json           # Postman automated tests
├── go.mod / go.sum             # Go (lang) dependency manifest
└── .gitignore
```

---


## License & Disclaimer

Built for **learning purposes only** — mock battle-tested for production. Use at your own risk.
Just use dawg.

— Made with 💙 by [@richky_4srg](https://www.instagram.com/richky_4srg/)
