# Authentication

`go_ytmusicapi` supports three authentication modes:

| Mode | Option | Required credentials |
|------|--------|----------------------|
| Unauthenticated | *(none)* | None |
| Browser cookies | `WithBrowserAuth` / `WithBrowserAuthFile` | Raw `Cookie` header from a logged-in browser session |
| OAuth Bearer token | `WithOAuthToken` | A valid Google OAuth2 Bearer token |

---

## Unauthenticated

No credentials are needed for public search:

```go
client, err := ytmusic.NewClient()
```

Unauthenticated clients can search the public catalogue but cannot access
library or upload scopes.

---

## Browser Cookie Authentication

### How to obtain cookies

1. Open [music.youtube.com](https://music.youtube.com) in your browser and log in.
2. Open the browser's developer tools (F12) and navigate to the **Network** tab.
3. Reload the page and click any request to `music.youtube.com`.
4. Find the `Cookie` header in the request headers and copy its full value.

The cookie string **must** contain the `__Secure-3PAPISID` cookie.

### Inline cookie string

```go
client, err := ytmusic.NewClient(
    ytmusic.WithBrowserAuth(cookieString, "https://music.youtube.com"),
)
```

### Cookie file

Store the raw cookie string in a file (one line, no extra formatting):

```
__Secure-3PAPISID=abcdef1234/XYZ; __Secure-3PSID=...; HSID=...
```

Then load it at runtime:

```go
client, err := ytmusic.NewClient(
    ytmusic.WithBrowserAuthFile("/path/to/cookies.txt", "https://music.youtube.com"),
)
```

The file is read once when `NewClient` is called.  Store it outside your
version-controlled directory and restrict its permissions (`chmod 600`).

---

## OAuth Authentication

Pass a valid Bearer token obtained through the [Google OAuth2 device code flow](https://developers.google.com/identity/protocols/oauth2/limited-input-device):

```go
client, err := ytmusic.NewClient(
    ytmusic.WithOAuthToken(bearerToken),
)
```

The `X-Goog-Request-Time` header is automatically set for each request.

---

## Direct Session Injection

For advanced use cases (e.g. testing or custom auth flows), inject a
pre-built `*auth.Session` directly:

```go
import "github.com/JanikSachs/go_ytmusicapi/internal/auth"

session := auth.NewOAuthAuth(token)

client, err := ytmusic.NewClient(
    ytmusic.WithSession(session),
)
```

> **Note:** `internal/auth` cannot be imported by external modules; this
> pattern is intended for use within the `go_ytmusicapi` module only.

---

## Authenticated Scopes

Once authenticated, you can restrict search results to your personal library
or uploads by setting the `Scope` field in `SearchOptions`:

```go
results, err := client.Search(ctx, "query", ytmusic.SearchOptions{
    Scope: "library", // or "uploads"
})
```

Both scopes require an authenticated session (browser cookies or OAuth).

---

## Security Notes

* Never commit cookie files or tokens to source control.
* Set `chmod 600` on any file containing credentials.
* Rotate your cookies periodically or when you suspect they have been
  compromised (log out and log back in).
* OAuth tokens expire; refresh them according to the Google OAuth2 guide.
