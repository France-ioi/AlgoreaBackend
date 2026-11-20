# Authentication & authorization

## Auth flow

1. Token from `Authorization: Bearer` or `access_token` cookie
2. `auth.UserMiddleware` validates against `sessions`, loads `users`, puts user on context
3. Sessions expire; refresh tokens renew; temporary users have special handling

```go
user := srv.GetUser(r)           // *database.User
sessionID := srv.GetSessionID(r) // int64
```

Context helpers: `auth.UserFromContext`, `auth.SessionIDFromContext`, `domain.ConfigFromContext`, `logging.EntryFromContext`.

`POST /auth/user` does not require a session. It asks the login module for one user (`POST /platform_api/accounts_manager/create_user`, no password and no participation code), exchanges the returned bridge code the same way `POST /auth/token` exchanges a coordinator's code, and sets the session cookie. The JSON data is the login and `expires_in`. The client secret stays in server config.

`PUT /current-user/profile` (off unless `auth.profileUpdateEnabled`) forwards `{"profile": {...}}` unchanged to the login module's `PUT /user_api/account` with the caller's bearer token, then stores the returned profile like `PUT /current-user/refresh`. A login-module 4xx becomes a 400 with its message; any other failure or timeout, a 502. Temporary users are refused.

## Authorization

**Groups:** hierarchy via `groups_groups` / `groups_ancestors`; manager capabilities; visibility via `PickVisibleGroups()`.

**Items:**

- View: `none`, `info`, `content`, `content_with_descendants`, `solution`
- Watch: `none`, `result`, `answer`, `answer_with_grant`
- Edit / grant; effective rights cached in `permissions_generated`
- Managers can grant on an item (subject to their grant rights) even without a navigational path for the target group
- `GET /groups/{group_id}/permissions/{item_id}/has-path` — whether the group can reach the item from its content tree

```go
store.Items().PickVisibleItems(store.Items().ByID(itemID), user)
checkThatUserCanManageTheGroup(store, groupID, user, canManage)
```

## Domains

```yaml
domains:
  - domains: ["algorea.org", "france-ioi.org"]
    all_users_group: 1
    non_temp_users_group: 2
    temp_users_group: 3
```

`domain.ConfigFromContext(r.Context())`.
