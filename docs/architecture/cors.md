# CORS

Configured in `app/cors.go` from the top-level `cors` section.

## `allowedOrigins`

Full origins (`scheme://host[:port]`). Supports bare `*` and per-host wildcards (`https://*.example.com`).

When unset or `[]`, `resolveAllowedOrigins` substitutes the `["none"]` sentinel before `go-chi/cors`. That sentinel is never a real browser `Origin`, so the library stays out of `allowedOriginsAll` (`len == 0`) and the default is **fail-closed** for every cross-origin request. List trusted origins explicitly to open access.

`resolveAllowedOrigins` splits on commas, trims, drops empties — so `ALGOREA_CORS__ALLOWEDORIGINS=https://a.example,https://b.example` matches a YAML list (viper v1.3.1 would otherwise leave one glued entry). Same split runs on YAML; a typo that puts two origins in one quoted string expands the trusted set. Prefer one `- https://…` per YAML entry.

## `allowCredentials`

Default `false`. When `true`, `allowedOrigins` must be a non-empty explicit list without bare `"*"`. Startup rejects (`errCORSCredentialsRequireExplicitOrigins`):

- Unset/empty origins (would be `["none"]`) — almost always an operator mistake with credentials enabled
- Bare `"*"` with credentials — CSRF-via-CORS

Per-host wildcards remain allowed with credentials.

## Migration note

Older behaviour hardcoded or fell back to `AllowedOrigins: ["*"]` (sometimes with credentials). Unset now always resolves to `["none"]` + `allowCredentials=false`. Local frontends that relied on echoed arbitrary origins need an explicit `cors:` block. In-process BDD does not exercise CORS.
