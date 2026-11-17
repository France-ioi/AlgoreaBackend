# Algorea Backend

API and persistence for the Algorea platform: users, groups, items, attempts, and permissions.

## Language

**User**:
A person on the platform, represented by a `User`-type group and a row in `users`.
_Avoid_: account, member (unless meaning group membership)

**Profile**:
The login-module `profile` JSON stored on the user (`users.profile`). It is not the flat personal columns on `users`.
_Avoid_: user info, personal info, account details

**Login**:
The name of a user on the login-module.
_Avoid_: username, account

**Login id**:
The login-module's identifier for a user. It is not the user's group id on this platform.
_Avoid_: user id

**Personal info**:
The user's identifying attributes (notably real name) whose visibility to others is governed by personal-info access approvals.
_Avoid_: profile, private data
