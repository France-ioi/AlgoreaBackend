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
