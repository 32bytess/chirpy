# Chirpy API Reference

Chirpy is a small social API: users sign up, log in, and post short messages called *chirps*.
This page documents every HTTP route registered in [`main.go`](../main.go), as currently implemented.

- [Overview](#overview)
- [Configuration](#configuration)
- [Authentication](#authentication)
- [Resources](#resources)
- [Endpoint summary](#endpoint-summary)
- [Health](#health)
- [Users](#users)
- [Tokens](#tokens)
- [Chirps](#chirps)
- [Webhooks](#webhooks)
- [Admin](#admin)
- [Static app](#static-app)
- [Known quirks](#known-quirks)

---

## Overview

| | |
|---|---|
| Base URL | `http://localhost:$PORT` |
| Request bodies | JSON (`Content-Type: application/json`) |
| Response bodies | JSON, except `/api/healthz` (plain text), `/admin/metrics` (HTML) and `/app/*` (static files) |
| IDs | UUID strings |
| Timestamps | RFC 3339 strings, e.g. `"2026-11-15T10:04:05.123456Z"` |

### Errors

Every error response has this shape:

```json
{ "error": "Human-readable message" }
```

The messages listed under each endpoint below are the exact strings the server returns.

### Empty responses

Responses with status `204 No Content` have no body.

---

## Configuration

The server reads these environment variables (a `.env` file in the working directory is loaded if present):

| Variable | Purpose |
|---|---|
| `DB_URL` | Postgres connection string |
| `PORT` | Port to listen on |
| `TOKEN_SECRET` | HMAC secret used to sign and verify access tokens (JWTs) |
| `POLKA_KEY` | API key that the Polka webhook must present |
| `PLATFORM` | Set to `dev` to enable `POST /admin/reset` |

---

## Authentication

Chirpy uses three kinds of credentials, all sent in the `Authorization` header.

### Access token (JWT)

```
Authorization: Bearer <access_token>
```

- Returned by `POST /api/login` (field `token`) and `POST /api/refresh`.
- HS256-signed JWT, issuer `chirpy-access`, subject = the user's ID.
- **Valid for 1 hour.**
- Required by: `PUT /api/users`, `POST /api/chirps`, `DELETE /api/chirps/{chirpID}`.
- A missing, malformed, expired or invalid token on these routes returns `401` with `{"error": "Unauthorized"}`.

The header must be exactly two space-separated words, the first being `Bearer`.

### Refresh token

```
Authorization: Bearer <refresh_token>
```

- Returned by `POST /api/login` (field `refresh_token`).
- 64-character hex string, stored in the database.
- **Valid for 60 days** unless revoked.
- Used only by `POST /api/refresh` (to get a new access token) and `POST /api/revoke`.

### Polka API key

```
Authorization: ApiKey <POLKA_KEY>
```

Used only by `POST /api/polka/webhooks`.

---

## Resources

### User

```json
{
  "id": "6f1b6a2e-6a5e-4a8b-9c1d-2f0e5b7d8c9a",
  "created_at": "2026-11-15T10:00:00.000000Z",
  "updated_at": "2026-11-15T10:00:00.000000Z",
  "email": "walt@example.com",
  "is_chirpy_red": false
}
```

| Field | Type | Notes |
|---|---|---|
| `id` | UUID | |
| `created_at` | timestamp | |
| `updated_at` | timestamp | |
| `email` | string | Unique across users |
| `is_chirpy_red` | boolean | `true` once the user has been upgraded via the Polka webhook |

The password is never returned.

### User with tokens

Returned only by `POST /api/login`: a [User](#user) plus two extra fields.

```json
{
  "id": "6f1b6a2e-6a5e-4a8b-9c1d-2f0e5b7d8c9a",
  "created_at": "2026-11-15T10:00:00.000000Z",
  "updated_at": "2026-11-15T10:00:00.000000Z",
  "email": "walt@example.com",
  "is_chirpy_red": false,
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "3f9c2b...e81a"
}
```

| Field | Type | Notes |
|---|---|---|
| `token` | string | Access token (JWT), valid 1 hour |
| `refresh_token` | string | Refresh token, valid 60 days |

### Chirp

```json
{
  "id": "0b7e4f3c-1d2a-4e5f-8a9b-c0d1e2f3a4b5",
  "created_at": "2026-11-15T10:05:00.000000Z",
  "updated_at": "2026-11-15T10:05:00.000000Z",
  "body": "Hello, world!",
  "user_id": "6f1b6a2e-6a5e-4a8b-9c1d-2f0e5b7d8c9a"
}
```

| Field | Type | Notes |
|---|---|---|
| `id` | UUID | |
| `created_at` | timestamp | |
| `updated_at` | timestamp | |
| `body` | string | At most 140 characters, after profanity filtering |
| `user_id` | UUID | The author |

Deleting a user also deletes their chirps and refresh tokens.

---

## Endpoint summary

| Method | Path | Auth | Success |
|---|---|---|---|
| `GET` | [`/api/healthz`](#get-apihealthz) | none | `200` |
| `POST` | [`/api/users`](#post-apiusers) | none | `201` |
| `PUT` | [`/api/users`](#put-apiusers) | access token | `200` |
| `POST` | [`/api/login`](#post-apilogin) | none | `200` |
| `POST` | [`/api/refresh`](#post-apirefresh) | refresh token | `200` |
| `POST` | [`/api/revoke`](#post-apirevoke) | refresh token | `204` |
| `POST` | [`/api/chirps`](#post-apichirps) | access token | `201` |
| `GET` | [`/api/chirps`](#get-apichirps) | none | `200` |
| `GET` | [`/api/chirps/{chirpID}`](#get-apichirpschirpid) | none | `200` |
| `DELETE` | [`/api/chirps/{chirpID}`](#delete-apichirpschirpid) | access token | `204` |
| `POST` | [`/api/polka/webhooks`](#post-apipolkawebhooks) | Polka API key | `204` |
| `GET` | [`/admin/metrics`](#get-adminmetrics) | none | `200` |
| `POST` | [`/admin/reset`](#post-adminreset) | none (`PLATFORM=dev` only) | `200` |
| `GET` | [`/app/*`](#get-app) | none | `200` |

A request to a known path with the wrong method gets `405 Method Not Allowed` from Go's router.

---

## Health

### `GET /api/healthz`

Readiness check.

**Response `200`** — `Content-Type: text/plain; charset=utf-8`

```
OK
```

```sh
curl -i http://localhost:8080/api/healthz
```

---

## Users

### `POST /api/users`

Create an account.

**Request body**

| Field | Type | Required |
|---|---|---|
| `email` | string | yes, non-empty |
| `password` | string | yes, non-empty |

```json
{ "email": "walt@example.com", "password": "04234" }
```

The password is stored as an Argon2id hash.

**Response `201`** — a [User](#user).

**Errors**

| Status | `error` | When |
|---|---|---|
| `400` | `Couldn't decode params` | Body is not valid JSON |
| `400` | `Bad Request` | `email` or `password` missing or empty |
| `400` | `User didnt get created` | Database insert failed, e.g. the email is already taken |
| `500` | `Something went wrong` | Password hashing failed |

```sh
curl -X POST http://localhost:8080/api/users \
  -H 'Content-Type: application/json' \
  -d '{"email":"walt@example.com","password":"04234"}'
```

### `PUT /api/users`

Change the authenticated user's email and password. Both fields are required; there is no partial update.

**Headers** — `Authorization: Bearer <access_token>`

**Request body**

| Field | Type | Required |
|---|---|---|
| `email` | string | yes, non-empty |
| `password` | string | yes, non-empty |

```json
{ "email": "walter@example.com", "password": "newpassword" }
```

**Response `200`** — the updated [User](#user).

**Errors**

| Status | `error` | When |
|---|---|---|
| `401` | `Unauthorized` | Missing or invalid access token |
| `400` | `Couldn't decode params` | Body is not valid JSON |
| `400` | `Unvalid inputs` | `email` or `password` missing or empty |
| `500` | `Something went wrong` | Hashing or database update failed (including the new email already belonging to another user) |

```sh
curl -X PUT http://localhost:8080/api/users \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"email":"walter@example.com","password":"newpassword"}'
```

### `POST /api/login`

Log in with email and password. Returns an access token and a new refresh token.

**Request body**

| Field | Type | Required |
|---|---|---|
| `email` | string | yes, non-empty |
| `password` | string | yes, non-empty |

```json
{ "email": "walt@example.com", "password": "04234" }
```

**Response `200`** — a [User with tokens](#user-with-tokens).

**Errors**

| Status | `error` | When |
|---|---|---|
| `400` | `Couldn't decode params` | Body is not valid JSON |
| `400` | `Unvalid inputs` | `email` or `password` missing or empty |
| `404` | `User not found` | No user has this email |
| `401` | `Unvalid Password` | Password does not match |
| `500` | `Something went wrong` | Hash check, token signing or refresh-token insert failed |

```sh
curl -X POST http://localhost:8080/api/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"walt@example.com","password":"04234"}'
```

---

## Tokens

### `POST /api/refresh`

Exchange a refresh token for a new access token (valid 1 hour). The refresh token itself is unchanged. No request body.

**Headers** — `Authorization: Bearer <refresh_token>`

**Response `200`**

```json
{ "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." }
```

**Errors**

| Status | `error` | When |
|---|---|---|
| `401` | `Invalid token` | `Authorization` header missing or not `Bearer <token>` |
| `500` | `Something went wrong` | Refresh token not found (see [Known quirks](#known-quirks)), or signing failed |
| `401` | `Unauthorized` | Refresh token has expired or been revoked |

```sh
curl -X POST http://localhost:8080/api/refresh \
  -H "Authorization: Bearer $REFRESH_TOKEN"
```

### `POST /api/revoke`

Revoke a refresh token (e.g. on logout). Afterwards `POST /api/refresh` rejects it with `401`. No request body.

**Headers** — `Authorization: Bearer <refresh_token>`

**Response `204`** — no body. Also returned when the token does not exist.

**Errors**

| Status | `error` | When |
|---|---|---|
| `400` | `Unvalid token` | `Authorization` header missing or not `Bearer <token>`, or the database update failed |

```sh
curl -X POST http://localhost:8080/api/revoke \
  -H "Authorization: Bearer $REFRESH_TOKEN"
```

---

## Chirps

### `POST /api/chirps`

Post a chirp as the authenticated user.

**Headers** — `Authorization: Bearer <access_token>`

**Request body**

| Field | Type | Required |
|---|---|---|
| `body` | string | yes, 1–140 bytes |

```json
{ "body": "I had something interesting for breakfast" }
```

**Profanity filter.** The body is split on whitespace, and any word equal to `kerfuffle`, `sharbert` or `fornax` (any case) is replaced with `****`. Words are rejoined with single spaces, so runs of whitespace and newlines collapse. A word with punctuation attached (e.g. `Fornax!`) is not replaced.

```
"This is a kerfuffle opinion"  →  "This is a **** opinion"
```

**Response `201`** — the created [Chirp](#chirp).

**Errors**

| Status | `error` | When |
|---|---|---|
| `401` | `Unauthorized` | Missing or invalid access token |
| `400` | `Couldn't decode params` | Body is not valid JSON |
| `400` | `body can't be nil` | `body` missing or empty |
| `400` | `Chirp is too long` | `body` longer than 140 bytes (checked before filtering) |
| `400` | `Chirp didnt get created` | Database insert failed |

```sh
curl -X POST http://localhost:8080/api/chirps \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"body":"Hello, world!"}'
```

### `GET /api/chirps`

List chirps, oldest first (`created_at` ascending).

**Query parameters**

| Name | Type | Required | Description |
|---|---|---|---|
| `author_id` | UUID | no | Only return chirps by this user |

**Response `200`** — an array of [Chirps](#chirp). An empty result is `[]`.

**Errors**

| Status | `error` | When |
|---|---|---|
| `400` | `Invalid User ID` | `author_id` is not a valid UUID |
| `500` | `Something went wrong` | Database query failed |

```sh
curl http://localhost:8080/api/chirps
curl "http://localhost:8080/api/chirps?author_id=$USER_ID"
```

### `GET /api/chirps/{chirpID}`

Get one chirp.

**Path parameters**

| Name | Type |
|---|---|
| `chirpID` | UUID |

**Response `200`** — a [Chirp](#chirp).

**Errors**

| Status | `error` | When |
|---|---|---|
| `404` | `Not found` | `chirpID` is not a valid UUID, or no chirp has this ID |

```sh
curl http://localhost:8080/api/chirps/$CHIRP_ID
```

### `DELETE /api/chirps/{chirpID}`

Delete a chirp. Only its author can delete it.

**Headers** — `Authorization: Bearer <access_token>`

**Path parameters**

| Name | Type |
|---|---|
| `chirpID` | UUID |

**Response `204`** — no body.

**Errors**

| Status | `error` | When |
|---|---|---|
| `401` | `Unauthorized` | Missing or invalid access token |
| `404` | `Not found` | `chirpID` is not a valid UUID, or no chirp has this ID |
| `403` | `Only the Owner can delete the chirp` | The chirp belongs to another user |
| `500` | `Something went wrong` | Database delete failed |

```sh
curl -X DELETE http://localhost:8080/api/chirps/$CHIRP_ID \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

---

## Webhooks

### `POST /api/polka/webhooks`

Called by the Polka payment provider. A `user.upgraded` event sets the user's `is_chirpy_red` to `true`. Every other event is acknowledged and ignored.

**Headers** — `Authorization: ApiKey <POLKA_KEY>`

**Request body**

| Field | Type | Description |
|---|---|---|
| `event` | string | Event name; only `user.upgraded` has an effect |
| `data.user_id` | UUID string | User to upgrade |

```json
{
  "event": "user.upgraded",
  "data": { "user_id": "6f1b6a2e-6a5e-4a8b-9c1d-2f0e5b7d8c9a" }
}
```

**Response `204`** — no body. Returned both when the user was upgraded and when the event was ignored.

**Errors** (checked in this order)

| Status | `error` | When |
|---|---|---|
| `401` | `Unauthorized` | `Authorization` header missing or not `ApiKey <key>` |
| `401` | `Invalid api key` | Key does not match `POLKA_KEY` |
| `400` | `Invalid inputs ` | Body is not valid JSON (message has a trailing space) |
| `400` | `Unvalid user_id` | Event is `user.upgraded` and `data.user_id` is not a valid UUID |
| `500` | `Something went wrong` | Database update failed |
| `404` | `User not found` | No user has this ID |

```sh
curl -X POST http://localhost:8080/api/polka/webhooks \
  -H "Authorization: ApiKey $POLKA_KEY" \
  -H 'Content-Type: application/json' \
  -d '{"event":"user.upgraded","data":{"user_id":"'"$USER_ID"'"}}'
```

---

## Admin

### `GET /admin/metrics`

Show how many requests the static app under `/app/` has served since start-up or the last reset.

**Response `200`** — `Content-Type: text/html`

```html
<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited 42 times!</p>
  </body>
</html>
```

```sh
curl http://localhost:8080/admin/metrics
```

### `POST /admin/reset`

Development only. Resets the hit counter to 0 and **deletes every user**, which also deletes all chirps and refresh tokens. No request body.

**Response `200`**

```json
{ "message": "Hits reset to 0 and delete all users" }
```

**Errors**

| Status | `error` | When |
|---|---|---|
| `403` | `403 Forbidden` | `PLATFORM` is not `dev` |
| `500` | `Something went wrong` | Database delete failed |

```sh
curl -X POST http://localhost:8080/admin/reset
```

---

## Static app

### `GET /app/*`

Serves files from the `app/` directory (`/app/` → `app/index.html`, `/app/assets/logo.png` → `app/assets/logo.png`). Each request increments the counter shown by `/admin/metrics`.

```sh
curl http://localhost:8080/app/
```

---

## Known quirks

Current behavior that clients may trip over. These are documented as-is, not fixed.

- **Unknown refresh token → `500`.** `POST /api/refresh` with a token that was never issued returns `500 Something went wrong` rather than `401`.
- **Revoking an unknown token succeeds.** `POST /api/revoke` returns `204` even if the token does not exist.
- **Login reveals whether an email is registered.** An unknown email returns `404 User not found`; a wrong password returns `401 Unvalid Password`.
- **Chirp length is counted in bytes**, so a chirp with multi-byte characters (accents, emoji) can be rejected below 140 visible characters.
- **Duplicate email** gives `400` on `POST /api/users` but `500` on `PUT /api/users`.
