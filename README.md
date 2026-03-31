# go_ytmusicapi

An unofficial Go client for the [YouTube Music](https://music.youtube.com) API.
This is a port of the Python [ytmusicapi](https://github.com/sigma67/ytmusicapi) project,
rewritten from scratch in idiomatic Go 1.24+.

## Features

| Feature | Status |
|---|---|
| Unauthenticated search | ✅ |
| Browser cookie authentication | ✅ |
| OAuth Bearer token authentication | ✅ |
| Search (songs, videos, albums, artists, playlists, profiles, podcasts, episodes) | ✅ |
| Artist lookup | 🚧 planned |
| Album lookup | 🚧 planned |
| Playlist lookup | 🚧 planned |
| Song metadata | 🚧 planned |

## Installation

```bash
go get github.com/JanikSachs/go_ytmusicapi
```

Requires **Go 1.24+**. No external dependencies — stdlib only.

## Quick Start

### Unauthenticated search

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/JanikSachs/go_ytmusicapi/pkg/ytmusic"
)

func main() {
    client, err := ytmusic.NewClient()
    if err != nil {
        log.Fatal(err)
    }

    results, err := client.Search(context.Background(), "Oasis Wonderwall")
    if err != nil {
        log.Fatal(err)
    }

    for _, r := range results {
        fmt.Printf("[%s] %s\n", r.ResultType, r.Title)
    }
}
```

### Filtered search

```go
results, err := client.Search(ctx, "Bohemian Rhapsody", ytmusic.SearchOptions{
    Filter: "songs",
    Limit:  10,
})
```

### Authenticated search (browser cookies)

```go
client, err := ytmusic.NewClient(
    ytmusic.WithBrowserAuth(cookieString, "https://music.youtube.com"),
)
```

### Authentication from cookie file

```go
client, err := ytmusic.NewClient(
    ytmusic.WithBrowserAuthFile("/path/to/cookies.txt", "https://music.youtube.com"),
)
```

### OAuth authentication

```go
client, err := ytmusic.NewClient(
    ytmusic.WithOAuthToken(bearerToken),
)
```

### Custom timeout and headers

```go
client, err := ytmusic.NewClient(
    ytmusic.WithTimeout(10 * time.Second),
    ytmusic.WithHeaders(map[string]string{"X-Custom": "value"}),
)
```

### Retry on transient errors

```go
client, err := ytmusic.NewClient(
    ytmusic.WithRetry(3, 500*time.Millisecond),
)
```

## Authentication

**Unauthenticated** requests work for public search results.

**Browser authentication** requires extracting cookie headers from a logged-in
YouTube Music browser session. The `__Secure-3PAPISID` cookie must be present.
You can pass the cookie string directly (`WithBrowserAuth`) or load it from a
file (`WithBrowserAuthFile`).

**OAuth** requires a valid Bearer token obtained through the Google OAuth2 device
code flow.

See [docs/auth.md](docs/auth.md) for detailed authentication guidance and
[docs/migration.md](docs/migration.md) for how the Python authentication
types map to Go equivalents.

## Configuration Options

| Option | Description | Default |
|---|---|---|
| `WithLanguage(lang)` | Response language (ISO code) | `"en"` |
| `WithLocation(loc)` | Geographic location (ISO country code) | `""` |
| `WithUserID(id)` | Brand account user ID | `""` |
| `WithHTTPClient(h)` | Custom `*http.Client` | 30 s timeout |
| `WithTimeout(d)` | Request timeout (overrides `WithHTTPClient` timeout) | `30s` |
| `WithHeaders(h)` | Extra HTTP headers added to every request | — |
| `WithBrowserAuth(cookie, origin)` | Browser cookie auth | — |
| `WithBrowserAuthFile(path, origin)` | Browser cookie auth loaded from file | — |
| `WithOAuthToken(token)` | OAuth Bearer token auth | — |
| `WithRetry(maxAttempts, wait)` | Retry on transient errors (5xx / network) | disabled |

## Search Filters

The `Filter` field in `SearchOptions` accepts:

`songs` · `videos` · `albums` · `artists` · `playlists` · `community_playlists` ·
`featured_playlists` · `profiles` · `podcasts` · `episodes`

Leave empty for a mixed search across all types.

## Search Scopes

The `Scope` field accepts `"library"` or `"uploads"` to restrict results
to the authenticated user's content. Authentication is required for scoped searches.

## Error Handling

```go
results, err := client.Search(ctx, query)
if err != nil {
    var serverErr *ytmusic.ServerError
    if errors.As(err, &serverErr) {
        fmt.Println("HTTP error:", serverErr.StatusCode)
    }
    // or check with errors.Is(err, ytmusic.ErrServer)
}
```

## Development

### Building

```bash
go build ./...
```

### Testing

```bash
go test ./...
```

### Linting / vetting

```bash
go vet ./...
```

### Running the example

```bash
go run ./cmd/example
```

## Architecture

See [docs/architecture.md](docs/architecture.md) for a full description of the
package layout and design decisions.

## Migration from Python ytmusicapi

See [docs/migration.md](docs/migration.md) for a mapping from the Python
`ytmusicapi` patterns to idiomatic Go equivalents.

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE) for details.
