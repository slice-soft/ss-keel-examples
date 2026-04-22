# 12 — OAuth2

OAuth2 login with Google, GitHub, and GitLab using [ss-keel-oauth](https://github.com/slice-soft/ss-keel-oauth) and [ss-keel-jwt](https://github.com/slice-soft/ss-keel-jwt).

> **Note:** This example uses a compact flat layout (`main.go` at the root) to keep all the relevant code in one place and let you focus on the API surface — no CLI scaffolding required. Projects created with `keel new` place the entry point at `cmd/main.go` and split the OAuth and JWT setup into dedicated `cmd/setup_oauth.go` and `cmd/setup_jwt.go` provider files.

## What This Demonstrates

- Initializing `*jwt.JWT` and passing it to `oauth.New(...)` as `contracts.TokenSigner`
- Enabling only specific providers via `OAUTH_ENABLED_PROVIDERS`
- Both delivery modes: JSON response and backend-to-frontend redirect
- Reading JWT claims on a protected `/api/me` route after login
- Building callback URLs from `OAUTH_REDIRECT_BASE_URL` + route prefix

## Requirements

- Go 1.25+
- At least one OAuth app registered on a supported provider (see below)

## How to Run

```bash
cp .env.example .env   # then fill in your provider credentials
go mod download
go run main.go
```

The server starts on port **7331**.

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/auth/google` | Redirect to Google login |
| GET | `/auth/google/callback` | Exchange code, sign JWT |
| GET | `/auth/github` | Redirect to GitHub login |
| GET | `/auth/github/callback` | Exchange code, sign JWT |
| GET | `/auth/gitlab` | Redirect to GitLab login |
| GET | `/auth/gitlab/callback` | Exchange code, sign JWT |
| GET | `/api/me` | Decode JWT claims (requires Bearer token) |
| GET | `/health` | Health check |
| GET | `/docs` | OpenAPI UI |

Only providers with complete credentials are registered. Providers with empty `ClientID` or `ClientSecret` are silently skipped.

## Provider Setup

| Provider | Credential source | Callback URL |
|----------|------------------|--------------|
| Google | [console.cloud.google.com → APIs & Services → Credentials](https://console.cloud.google.com/apis/credentials) | `http://127.0.0.1:7331/auth/google/callback` |
| GitHub | [github.com/settings/developers → OAuth Apps](https://github.com/settings/developers) | `http://127.0.0.1:7331/auth/github/callback` |
| GitLab | [gitlab.com/-/user_settings/applications](https://gitlab.com/-/user_settings/applications) | `http://127.0.0.1:7331/auth/gitlab/callback` |
