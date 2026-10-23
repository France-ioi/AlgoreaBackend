# API & service layer

## Packages

Under `app/api/`: `answers`, `auth`, `currentuser`, `groups`, `items`, `threads`, `users`. Each package:

1. `<resource>.go` — `Service` + `SetRoutes`
2. `<operation>.go` — one handler per file
3. `*.feature` / `bdd_test.go` — BDD

```go
type Service struct {
    *service.Base
}

func (srv *Service) SetRoutes(router chi.Router) {
    router.Use(render.SetContentType(render.ContentTypeJSON))
    router.Use(auth.UserMiddleware(srv.Base))
    router.Get("/groups/{group_id}", service.AppHandler(srv.getGroup).ServeHTTP)
}
```

## Handler shape

```go
func (srv *Service) getGroup(w http.ResponseWriter, r *http.Request) error {
    groupID, err := service.ResolveURLQueryPathInt64Field(r, "group_id")
    if err != nil {
        return service.ErrInvalidRequest(err)
    }
    user := srv.GetUser(r)
    store := srv.GetStore(r)

    var group GroupResponse
    err = store.Groups().PickVisibleGroups(store.Groups().ByID(groupID), user).
        Scan(&group).Error()
    if gorm.IsRecordNotFoundError(err) {
        return service.ErrAPIInsufficientAccessRights
    }
    service.MustNotBeError(err)

    render.Respond(w, r, group)
    return nil
}
```

## Errors & responses

`APIError` helpers in `service/errors.go`: `ErrInvalidRequest` (400), `ErrUnauthorized` (401), `ErrAPIInsufficientAccessRights` / `ErrForbidden` (403), `ErrNotFound` (404), `ErrRequestTimeout` (408), `ErrUnexpected` (500). `AppHandler` turns them into JSON.

`go-chi/render` + `service.AppResponder`: success gets `success: true`; errors `success: false` + message; pagination metadata when present.

## Middleware (app stack)

Logger → DataStore → VersionHeader → RealIP → Compression? → RequestID → StructuredLogger → Recoverer → CORS → Domain → (per-service) UserMiddleware.

## `service.Base`

```go
type Base struct {
    store             *database.DataStore
    ServerConfig      *viper.Viper
    AuthConfig        *viper.Viper
    PropagationConfig *viper.Viper
    DomainConfig      []domain.ConfigItem
    TokenConfig       *token.Config
}
```

`GetUser` / `GetSessionID` / `GetStore` from the request.

| Utility | File | Role |
|---------|------|------|
| Path/query parsers | `parameters.go` | `ResolveURLQueryPathInt64Field`, … |
| `QueryLimiter` | `query_limiter.go` | `?limit` / `?offset` |
| `SortBuilder` | `sorting.go` | Sort fields |
| `SchedulePropagation` | `propagation.go` | sync after commit vs `propagation_requested` event |
