# Python → Go Migration Guide

This document maps the Python `ytmusicapi` concepts to their Go equivalents in `go_ytmusicapi`.

## Creating a Client

**Python**
```python
from ytmusicapi import YTMusic
yt = YTMusic()                          # unauthenticated
yt = YTMusic("browser.json")            # browser auth
yt = YTMusic("oauth.json")              # OAuth
```

**Go**
```go
import "github.com/JanikSachs/go_ytmusicapi/pkg/ytmusic"

client, err := ytmusic.NewClient()      // unauthenticated

client, err := ytmusic.NewClient(
    ytmusic.WithBrowserAuth(cookieStr, origin),
)

client, err := ytmusic.NewClient(
    ytmusic.WithOAuthToken(bearerToken),
)
```

## Searching

**Python**
```python
results = yt.search("Bohemian Rhapsody", filter="songs", limit=5)
```

**Go**
```go
results, err := client.Search(ctx, "Bohemian Rhapsody", ytmusic.SearchOptions{
    Filter: "songs",
    Limit:  5,
})
```

### Valid Filters

| Python `filter=`        | Go `SearchOptions.Filter` |
|-------------------------|---------------------------|
| `"songs"`               | `"songs"`                 |
| `"videos"`              | `"videos"`                |
| `"albums"`              | `"albums"`                |
| `"artists"`             | `"artists"`               |
| `"playlists"`           | `"playlists"`             |
| `"community_playlists"` | `"community_playlists"`   |
| `"featured_playlists"`  | `"featured_playlists"`    |
| `"profiles"`            | `"profiles"`              |
| `"podcasts"`            | `"podcasts"`              |
| `"episodes"`            | `"episodes"`              |

### Valid Scopes

| Python `scope=`  | Go `SearchOptions.Scope` |
|------------------|--------------------------|
| `"library"`      | `"library"`              |
| `"uploads"`      | `"uploads"`              |

## Result Types

Python returns plain dicts. Go returns `[]*ytmusic.SearchResult` — a typed struct.

```go
for _, r := range results {
    fmt.Println(r.ResultType) // "song", "video", "album", etc.
    fmt.Println(r.Title)
    fmt.Println(r.VideoID)
    if r.DurationSec != nil {
        fmt.Printf("%d seconds\n", *r.DurationSec)
    }
}
```

## Client Options

| Python constructor arg / method | Go `Option`                          |
|---------------------------------|--------------------------------------|
| `language=`                     | `ytmusic.WithLanguage("en")`         |
| `location=`                     | `ytmusic.WithLocation("US")`         |
| `proxies=`                      | `ytmusic.WithHTTPClient(httpClient)` |
| `user_id=`                      | `ytmusic.WithUserID("...")`          |

## Error Handling

Python raises exceptions. Go returns errors:

```go
results, err := client.Search(ctx, query, opts)
if err != nil {
    var sErr *ytmusic.ServerError
    if errors.As(err, &sErr) {
        fmt.Printf("HTTP %d: %s\n", sErr.StatusCode, sErr.Message)
    }
    // or use sentinel errors:
    if errors.Is(err, ytmusic.ErrServer) { ... }
    if errors.Is(err, ytmusic.ErrUser)   { ... }
}
```

## Context / Cancellation

Go functions accept a `context.Context` as the first argument, enabling timeouts and cancellation — a feature absent from the Python library.

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
results, err := client.Search(ctx, "query")
```
