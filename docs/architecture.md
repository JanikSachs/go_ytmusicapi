# Architecture

## Overview

`go_ytmusicapi` is a Go 1.24 port of the [ytmusicapi](https://github.com/sigma67/ytmusicapi) Python library. It provides an idiomatic Go client for the unofficial YouTube Music web API.

## Package Layout

```
go_ytmusicapi/
├── cmd/
│   └── example/          # Runnable example binary
├── internal/
│   ├── auth/             # Authentication (browser cookies, OAuth)
│   ├── endpoints/        # Request body builders and parameter encoders
│   ├── model/            # Shared data model types
│   ├── parser/           # JSON navigation helpers and response parsers
│   └── transport/        # Low-level HTTP client wrapper
└── pkg/
    └── ytmusic/          # Public API surface (Client, options, types)
```

## Key Components

### `internal/transport`
Wraps `net/http` with YouTube Music-specific defaults: base URL, user-agent, SOCS cookie, and JSON marshalling/unmarshalling. Exposes `Post` and `Get` methods.

### `internal/auth`
Three authentication modes:
- **Unauthorized** – unauthenticated access (public search, etc.)
- **Browser** – cookie-based auth extracted from a browser session; generates `SAPISIDHASH` `Authorization` headers.
- **OAuth** – Bearer token auth.

### `internal/model`
Plain Go structs shared across internal packages (`SearchResult`, `Thumbnail`, `ArtistRef`, etc.). JSON tags match the Python library's output field names.

### `internal/parser`
- **`nav.go`** – `Nav` / `NavStr` / `NavList` / `NavMap`: type-safe path navigation over deeply nested `map[string]any` / `[]any` API responses.
- **`utils.go`** – `GetFlexColumnItem`, `GetItemText`, `ParseDuration`, `ParseIDName`.
- **`songs.go`** – `ParseSongRuns`: parses the flex-column "runs" that encode artist, album, duration, year, and views.
- **`search.go`** – `ParseSearchResult`, `ParseSearchResults`, `ParseTopResult`, `ParseThumbnails`, `ParseMenuPlaylists`.

### `internal/endpoints`
Builds the JSON request body and encodes the `params` query string for each API endpoint. Currently implements `search`.

### `pkg/ytmusic`
The public-facing package:
- **`Client`** – created via `NewClient(opts...)`.
- **`Option`** – functional options (`WithLanguage`, `WithLocation`, `WithBrowserAuth`, `WithOAuthToken`, `WithHTTPClient`, …).
- **`SearchOptions`** – per-call search configuration (filter, scope, limit, ignore-spelling).
- Type aliases re-exporting `internal/model` types so callers only import `pkg/ytmusic`.

## Data Flow

```
caller → Client.Search()
           │
           ├─ endpoints.GetSearchParams()  → encoded params string
           ├─ endpoints.SearchBody()       → JSON body map
           ├─ endpoints.Context()          → context body map
           │
           ├─ Client.sendRequest()         → HTTP POST to YTM API
           │       auth.Session.Headers()  → auth headers injected
           │
           └─ parseSearchResponse()
                   parser.ParseTopResult()      → top result card
                   parser.ParseSearchResults()  → shelf items
                       parser.ParseSearchResult()
                           parser.ParseSongRuns()
                           parser.ParseThumbnails()
```

## Error Handling

- `ServerError` (wraps `ErrServer`) – non-2xx API responses.
- `UserError` (wraps `ErrUser`) – invalid caller input (bad filter/scope).
- All errors are wrapped with `%w` for `errors.Is` / `errors.As` compatibility.
