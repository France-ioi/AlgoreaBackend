# Profile is the login-module JSON; three user fields stay on the wire as null

`GET /current-user` returns only `group_id`, `login`, `profile`, `default_language`, and `temp_user`. Profile is the login-module `profile` JSON stored on the user. Flat personal and preference columns were removed because clients no longer read them, and they overlapped Profile.

User `grade`, `free_text`, and `web_site` stay on the other user and member endpoints as permanent nulls so older clients keep parsing. They are not stored. `groups.grade` is a different attribute and stays. Flat email is not copied into Profile; email is read from Profile only when the login module puts it there.
