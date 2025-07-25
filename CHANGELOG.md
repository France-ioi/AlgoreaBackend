# Changelog
All notable changes to this project will be documented in this file.

## [v2.28.3](https://github.com/France-ioi/AlgoreaBackend/compare/v2.28.2...v2.28.3) - 2025-07-08
- fix and rework marking permissions for recomputing in triggers of `items_items`
- optimize permissions marking in `after_update_items_items` even more
- validate tokens returned by the login module
- enable more linters
- upgrade `golangci-lint` to v1.64.7
- optimize permissions marking in `before_delete_items_items` even more

## [v2.28.2](https://github.com/France-ioi/AlgoreaBackend/compare/v2.28.1...v2.28.2) - 2025-06-27
- speed up `groupRootsView` even more
- make the `resultStart` service less locking
- return 404 and delete the session in `refreshAccessToken` when the login module does not recognize the refresh token
- never fall back to sync propagations
- do not call the propagation endpoint inside transactions
- restore more linters
- remove outdated commands
- refactor internals

## [v2.28.1](https://github.com/France-ioi/AlgoreaBackend/compare/v2.28.0...2.28.1) - 2025-06-12
- temporarily disable after_insert_groups_groups trigger when parent_group_id=4 (non-temp user group) in order to speed up the migration on large db (prod)

## [v2.28.0](https://github.com/France-ioi/AlgoreaBackend/compare/v2.27.2...2.28.0) - 2025-06-06
- add support for the NonTempUsers group in the code, add the nonTempUsers group config parameters, move non-temporary users into this group from AllUsers

## [v2.27.2](https://github.com/France-ioi/AlgoreaBackend/compare/v2.27.1...2.27.2) - 2025-06-06
- create non-temp user group in preparation for the next version

## [v2.27.1](https://github.com/France-ioi/AlgoreaBackend/compare/v2.27.0...2.27.1) - 2025-06-04
- lock the `groups_groups` row `FOR UPDATE` when checking if a user is already a member of a badge group in GroupStore.storeBadge
- rework `itemPathFromRootFind` & `itemBreadcrumbsFromRootsGet`: prefer paths having attempts for final items + require items visibility correctly
- try to speed up the main SQL query of groupsRootView
- restore 'prealloc' linter, partially restore disabled linters + do not expose unexpected errors to end-clients

## [v2.27.0](https://github.com/France-ioi/AlgoreaBackend/compare/v2.26.4...v2.27.0) - 2025-05-15
- add (`can_enter_from`, `can_enter_until`) intervals in `itemView`, fix calculation of `can_request_help` for teams there
- `saveGrade`: handle the situation when `idItemLocal` of `score_token` is not given or is not a numeric string
- rework row locking in `groupManagerDelete`
- make the `can_manage` field in the output of `groupManagersView` null when a manager is not direct manager of the group
- fix the message about propagation in `db-recompute` command
- optimize the SQL query generated by `ItemStore.GetAncestorsRequestHelpPropagatedQuery()`
- get rid of the `N+1` selects problem in `permissionsView`
- `saveGrade`: fix the error message when `score_token` is given for a platform not having a public key + improve input checks

## [v2.26.4](https://github.com/France-ioi/AlgoreaBackend/compare/v2.26.3...v2.26.4) - 2025-04-30
- make `resultStartPath` and `itemCreate` less locking
- fix the result type of itemGetAdditionalTime in swagger docs
- fix dbdoc-gen on CircleCI

## [v2.26.3](https://github.com/France-ioi/AlgoreaBackend/compare/v2.26.2...v2.26.3) - 2025-03-27
- try closing the DB connection after AWS Lambda gets SIGTERM

## [v2.26.2](https://github.com/France-ioi/AlgoreaBackend/compare/v2.26.1...v2.26.2) - 2025-03-25
- use innodb_ft_user_stopword_table instead of innodb_ft_server_stopword_table for fulltext migration

## [v2.26.1](https://github.com/France-ioi/AlgoreaBackend/compare/v2.26.0...v2.26.1) - 2025-03-24
- fix fulltext search

## [v2.26.0](https://github.com/France-ioi/AlgoreaBackend/compare/v2.25.1...v2.26.0) - 2025-03-19
- upgrade MySQL to 8.0.34
- get rid of the term "contest" except for the mentioning of "contest participants groups"
- upgrade akrylysov/algnhsa to v1.1.0

## [v2.25.1](https://github.com/France-ioi/AlgoreaBackend/compare/v2.25.0...v2.25.1) - 2025-03-13
- convert all the db tables into the utf8mb4 charset and use it for MySQL connections
- set deleteWithTrapsBatchSize to 30 instead of 1000 (should decrease locking time during temp user deletion)
- ensure transactions in auth.CreateNewTempSession() and auth.RefreshTempUserSession()

## [v2.25.0](https://github.com/France-ioi/AlgoreaBackend/compare/v2.24.8...v2.25.0) - 2025-03-11
- implement a service for getting additional time of a group on a contest
- remove special characters from search strings in database.DB.WhereSearchStringMatches()
- docker improvements
- do not analyze tables or recompute db caches in db-migrate command + always close the DB connection in commands
- add a test checking that connection resetting restores the default value of FOREIGN_KEY_CHECKS setting
- rework app.Server.Start() to return errors instead of exiting the app + change the server's port in tests

## [v2.24.8](https://github.com/France-ioi/AlgoreaBackend/compare/v2.24.7...v2.24.8) - 2025-02-10
- acquire shared row locks instead of exclusive row locks in DB queries causing request timeouts in production
- add a comment for logRawSQLQueries in config sample files

## [v2.24.7](https://github.com/France-ioi/AlgoreaBackend/compare/v2.24.6...v2.24.7) - 2025-01-29
- fix MySQL triggers related to sync propagations

## [v2.24.6](https://github.com/France-ioi/AlgoreaBackend/compare/v2.24.5...v2.24.6) - 2025-01-24
- prevent mutual blocking of concurrent sync propagations

## [v2.24.5](https://github.com/France-ioi/AlgoreaBackend/compare/v2.24.4...v2.24.5) - 2025-01-23
- fix retrying on duplicate key errors, log such errors using INFO log level

## [v2.24.4](https://github.com/France-ioi/AlgoreaBackend/compare/v2.24.3...v2.24.4) - 2025-01-17
- bugfix: select the newly created/updated result FOR UPDATE in resultStart
- require the current user be able to view content of the item in order to modify threads in threadUpdate

## [v2.24.3](https://github.com/France-ioi/AlgoreaBackend/compare/v2.24.2...v2.24.3) - 2025-01-09
- adapt threads-related services to the recent forum permission rules change + fix some bugs related to permissions checking there
- handle DB errors during token unmarshalling correctly in `itemGetHintToken` & `saveGrade`
- print readable values of expected JSON rows in TheResponseAtShouldBe() when lengths do not match
- fix a mistake in a cucumber step name

## [v2.24.2](https://github.com/France-ioi/AlgoreaBackend/compare/v2.24.1...v2.24.2) - 2024-12-17
- add `user_id` into "request complete" logs

## [v2.24.1](https://github.com/France-ioi/AlgoreaBackend/compare/v2.24.0...v2.24.1) - 2024-12-13
- fix: render `item_id` of `unlocked_items` in `saveGrade` as string

## [v2.24.0](https://github.com/France-ioi/AlgoreaBackend/compare/v2.23.7...v2.24.0) - 2024-12-13
- return unlocked items in `saveGrade` + other fixes (major change to the propagation process)
- mark the title as nullable in swagger docs of `itemBreadcrumbsGet`
- eliminate data races when reading/setting hooks related to forceful retrying of transactions (used only in tests)
- make `Test_Deadline` stable

## [v2.23.7](https://github.com/France-ioi/AlgoreaBackend/compare/v2.23.6...v2.23.7) - 2024-11-27
- log retryable DB errors (deadlocks and lock wait timeouts) as INFO, rework logging, introduce console log formatter, add req_id into every log entry implicitly

## [v2.23.6](https://github.com/France-ioi/AlgoreaBackend/compare/v2.23.5...v2.23.6) - 2024-11-20
- make tests of app/logging cacheable
- use the latest version of Gorm from jinzhu/gorm instead of a patched version from France-ioi/gorm
- add even more output suppressing in tests
- bug fix: mark results of a parent item as 'to_be_recomputed' on items_items insertion
- fix itemAnswerGetResponse used by swagger docs of currentAnswerGet, answerGet, bestAnswerGet
- bug fix: use different named locks for the propagation command and for the results propagation

## [v2.23.5](https://github.com/France-ioi/AlgoreaBackend/compare/v2.23.4...v2.23.5) - 2024-11-18
- db wrappers: handle DB request timeouts
- log SQL queries everywhere
- analyze SQL statement
- rework named locking
- add 408 status code
- test-system improvements
- do not run groups ancestors propagation in currentUserDeletion & userBatchRemove services and delete-temp-users command
- schedule permissions propagations when needed in currentUserDeletion & userBatchRemove
- groupRemoveChild & groupDelete: mark permissions for propagation properly, run only needed propagations

## [v2.23.4](https://github.com/France-ioi/AlgoreaBackend/compare/v2.23.3...v2.23.4) - 2024-11-12
- make groups and items propagations faster
- fix/improve testing

## [v2.23.3](https://github.com/France-ioi/AlgoreaBackend/compare/v2.23.2...v2.23.3) - 2024-10-28
- make all the services compatible with transactions retrying
- optimize the groupChildrenView service
- fix/improve testing

## [v2.23.2](https://github.com/France-ioi/AlgoreaBackend/compare/v2.23.1...v2.23.2) - 2024-10-21
- rework groups propagation: get rid of the named lock, use more granular locking, speed up the propagation a bit

## [v2.23.1](https://github.com/France-ioi/AlgoreaBackend/compare/v2.23.0...v2.23.1) - 2024-10-17
- increase the timeout on item/group ancestor recomputation to prevent error 500 on the authentication service

## [v2.23.0](https://github.com/France-ioi/AlgoreaBackend/compare/v2.22.8...v2.23.0) - 2024-10-16
- speed up groupRootsView
- move ancestors recalculation back into initiating transactions + more granular locking during ancestors recalculation
- retry DB transactions on lock wait timeout errors (similarly to retrying on deadlocks)
- store the DB time instead of the server time in threads.latest_update_at and compare its value with the DB time instead of the server time as well
- do not close the response body in iSendrequestGeneric() as it is closed inside SendTestHTTPRequest()
- allow passing DB transaction options + pass the context for transactions #1186
- internal: introduce a new method database.DB.With(), rework database.DB.Union() & database.DB.UnionAll()
- internal: add missing call to testhelpers.SuppressOutputIfPasses() in integration tests
- rework time mocking in testhelpers

## [v2.22.8](https://github.com/France-ioi/AlgoreaBackend/compare/v2.22.7...v2.22.8) - 2024-10-01
- speed up the results propagation and make it less locking
- introduce a command recomputing all the results of chapters/skills

## [v2.22.7](https://github.com/France-ioi/AlgoreaBackend/compare/v2.22.6...v2.22.7) - 2024-09-25
- use even smaller iterations (200) for result propagation

## [v2.22.6](https://github.com/France-ioi/AlgoreaBackend/compare/v2.22.5...v2.22.6) - 2024-09-24
- use smaller iterations (1000) for result propagation
- rework processing of results_propagate_items table
- do not close the db connection explicitly in the propagation command

## [v2.22.5](https://github.com/France-ioi/AlgoreaBackend/compare/v2.22.4...v2.22.5) - 2024-09-19
- split the results propagation process into small atomic chunks

## [v2.22.4](https://github.com/France-ioi/AlgoreaBackend/compare/v2.22.3...v2.22.4) - 2024-09-16
- add a config to disable result propagation completely
- remove parameter to the propagation command to disable result propagation

## [v2.22.3](https://github.com/France-ioi/AlgoreaBackend/compare/v2.22.2...v2.22.3) - 2024-09-13
- add parameter to the propagation command to disable result propagation

## [v2.22.2](https://github.com/France-ioi/AlgoreaBackend/compare/v2.22.1...v2.22.2) - 2024-09-02
- add a delay parameter to delete-temp-users command

## [v2.22.1](https://github.com/France-ioi/AlgoreaBackend/compare/v2.22.0...v2.22.1) - 2024-08-26
- fix a bug: translation may be registered several times in a transaction, causing a crash

## [v2.22.0](https://github.com/France-ioi/AlgoreaBackend/compare/v2.21.4...v2.22.0) - 2024-08-21
- fix documentation generation, and the doc of several services:
  1. In responses of `invitationsView`, `inviting_user.first_name` & `inviting_user.last_name` become nullable (they should be as corresponding columns are nullable in the DB).
  2. In responses of `listThreads`, `participant.first_name` & `participant.last_name` are only shown if the current user has rights to view them (similarly to other services).
- change CORS to expose `Backend-Version` and `Date`, and hide `Link` headers
- return the modified result data (together with the linked attempt data) in the format of `attemptsList` response row on success in `resultStart`

## [v2.21.4](https://github.com/France-ioi/AlgoreaBackend/compare/v2.21.3...v2.21.4) - 2024-08-08
- speed up the `itemActivityLogForItem` service
- `itemActivityLogForItem`: fix `can_watch_answer` and how `can_watch_answer` is handled
- fix `generateProfileEditToken`, use loginIDs instead of group ids
- refactoring and internal improvements

## [v2.21.3](https://github.com/France-ioi/AlgoreaBackend/compare/v2.21.2...v2.21.3) - 2024-08-06
- fixing doc of `invitationsView` (after change from previous version)
- refactoring and internal improvements

## [v2.21.2](https://github.com/France-ioi/AlgoreaBackend/compare/v2.21.1...v2.21.2) - 2024-07-25
- update invitationsView doc as the invitation initiator may be null
- many refactoring and internal improvements

## [v2.21.1](https://github.com/France-ioi/AlgoreaBackend/compare/v2.21.0...v2.21.1) - 2024-07-23
- speed up the `invitationsView` service
- many refactoring and internal improvements

## [v2.21.0](https://github.com/France-ioi/AlgoreaBackend/compare/v2.20.1...v2.21.0) - 2024-07-09
- change of the `invitationsView` service API

## [v2.20.1](https://github.com/France-ioi/AlgoreaBackend/compare/v2.20.0...v2.20.1) - 2024-07-03
- use full-text indexes for item/group searches
- fix 'reinvite' after changing group approval

## [v2.20.0](https://github.com/France-ioi/AlgoreaBackend/compare/v2.19.0...v2.20.0) - 2024-06-14
- Service for getting a token for editing another user's profile
- userViewById: add whether we can view/edit his personnal info
- createAccessTokenToken: Limit number of session at 10 / users to avoid session spamming
- Delete expired token on session refresh
- Refactor existing and temp user session processes
- Internal: Add a stack trace in the logs when the binary crashes.
- Internal: Make sure the tests cannot be run on live database as it empties the database

## [v2.19.0](https://github.com/France-ioi/AlgoreaBackend/compare/v2.18.8...v2.19.0) - 2024-06-05
- allow not providing the access token when the task token is provided
- save grade service: use score/answer token as auth

## [v2.18.8](https://github.com/France-ioi/AlgoreaBackend/compare/v2.18.7...v2.18.8) - 2024-05-15
- make `startResultPath` and `startResult` use the async propagation
- bug fix: possible deadlock in propagation

## [v2.18.7](https://github.com/France-ioi/AlgoreaBackend/compare/v2.18.6...v2.18.7) - 2024-05-14
- speed up the "get-result-path" service

## [v2.18.6](https://github.com/France-ioi/AlgoreaBackend/compare/v2.18.5...v2.18.6) - 2024-05-13
- speed up the "list root activities" service

## [v2.18.5](https://github.com/France-ioi/AlgoreaBackend/compare/v2.18.4...v2.18.5) - 2024-05-09
- speed up item children edition

## [v2.18.4](https://github.com/France-ioi/AlgoreaBackend/compare/v2.18.3...v2.18.4) - 2024-05-08
- speed up permission propagation
- step-by-step propagation for propagations related to item ancestors and group ancestors

## [v2.18.3](https://github.com/France-ioi/AlgoreaBackend/compare/v2.18.2...v2.18.3) - 2024-04-26
- improve performance for auth
- add more logs to track db timings

## [v2.18.2](https://github.com/France-ioi/AlgoreaBackend/compare/v2.18.1...v2.18.2) - 2024-04-23
- improve performance for the start-result-path service

## [v2.18.1](https://github.com/France-ioi/AlgoreaBackend/compare/v2.18.0...v2.18.1) - 2024-04-22
- improve performance of the update item children service

## [v2.18.0](https://github.com/France-ioi/AlgoreaBackend/compare/v2.17.0...v2.18.0) - 2024-04-21
- fix nil pointer dereference when the schedule propagation endpoint call returns an error
- permission propagation is now split into smaller pieces to avoid timeouting
- all propagations are now scheduled so that they are run after the current transaction

## [v2.17.0](https://github.com/France-ioi/AlgoreaBackend/compare/v2.16.0...v2.17.0) - 2024-04-04
- update group service: handle change in approval policies

## [v2.16.0](https://github.com/France-ioi/AlgoreaBackend/compare/v2.15.0...v2.16.0) - 2024-03-14
- authMiddleware: explicitely disallow access if the token > max token size
- Get group service: add required approval info

## [v2.15.0](https://github.com/France-ioi/AlgoreaBackend/compare/v2.14.2...v2.15.0) - 2024-03-12
- itemActivityLogForItem & itemActivityLogForAllItems: add can_watch_item_answer in response: whether the current user can watch the answer
- add "isEmpty" info to group member services
- update sessions database schema & parallel session logout (internal changes, no change to API yet)
- internal improvements

## [v2.14.2](https://github.com/France-ioi/AlgoreaBackend/compare/v2.14.1...v2.14.2) - 2023-12-18
- make the request returning progresses of a group or user faster

## [v2.14.1](https://github.com/France-ioi/AlgoreaBackend/compare/v2.14.0...v2.14.1) - 2023-10-18
- fix a bug in the item children service related with skills

## [v2.14.0](https://github.com/France-ioi/AlgoreaBackend/compare/v1.13.2...v2.14.0) - 2023-10-04
- jump from v1.x.y to v2.x.y to match how we usually name this backend
- allow asynchronous permisssion and result propagation by calling an external endpoint
- fix: `getItem` service should return `can_request_help = true` in its permissions when the user is an owner

## [v1.13.2](https://github.com/France-ioi/AlgoreaBackend/compare/v1.13.1...v1.13.2) - 2023-09-25
- allow item owners to request help to any visible group
- rename attribute name related with the request help permission in the `getItem` service

## [v1.13.1](https://github.com/France-ioi/AlgoreaBackend/compare/v1.13.0...v1.13.1) - 2023-09-19
- small fixes, mainly to the doc, related with the thread services

## [v1.13.0](https://github.com/France-ioi/AlgoreaBackend/compare/v1.12.0...v1.13.0) - 2023-09-18
- add `can_request_help` information into the `getItem` service

## [v1.12.0](https://github.com/France-ioi/AlgoreaBackend/compare/v1.11.0...v1.12.0) - 2023-09-14
- fix services where ids were returned as numbers (instead of string)
- fix duplication in the thread listing service
- viewGrantedPermission: improve can_request_help_to support
- updatePermissions: allow non-visible can_request_help_to value if unchanged
- fix doc in general and for a few services

## [v1.11.0](https://github.com/France-ioi/AlgoreaBackend/compare/v1.10.0...v1.11.0) - 2023-09-07
- implement request-help permission propagation
- improve / fix the services to look for the path to some content, add info whether the path has been already been visited
- fix doc

## [v1.10.0](https://github.com/France-ioi/AlgoreaBackend/compare/v1.9.3...v1.10.0) - 2023-07-17
- get permission service: add `can_request_help_to` in granted permissions
- updatePermissions: allow updating `can_request_help_to`
- improve doc

## [v1.9.3](https://github.com/France-ioi/AlgoreaBackend/compare/v1.9.2...v1.9.3) - 2023-07-13
- add parameter to the token refresh service to allow to create or not a temp user on refresh failure

## [v1.9.2](https://github.com/France-ioi/AlgoreaBackend/compare/v1.9.1...v1.9.2) - 2023-07-12
- disable dynamic linking librairies in order to fix a deployment issue

## [v1.9.1](https://github.com/France-ioi/AlgoreaBackend/compare/v1.9.0...v1.9.1) - 2023-07-12
- minor fixes

## [v1.9.0](https://github.com/France-ioi/AlgoreaBackend/compare/v1.8.1...v1.9.0) - 2023-07-07

- forum: new thread listing service
- get-item service returns `description` for users with `can_view=info` perm level
- get best answer: distinguish "no answer" error from the access right errors
- get participant progress: do not return children if parents do not have results
- fix bug (crash) when setting a `root_skill_id` to `null` for a group
- add token to the get thread service
- hint request service: do not allow unsigned requests
- item navigation service: only return skills as children of skills
- inject backend version in responses
- get participant progress: add a `started_at` attribute
- access token create: create a temp user when no code provided and user is not authenticated (prevent 401 and so warning in browsers)
- path from root item: fix some bugs
- get best answer: return a success response when there is no answer (to prevent warning in browsers)
- get granted permissions: add `can_request_help_to` permission
- many code and test improvements
- upgrade to Go 1.20

## [v1.8.1](https://github.com/France-ioi/AlgoreaBackend/compare/v1.8.0...v1.8.1) - 2023-03-23

- fix swagger doc
- fix migrations

## [v1.8.0](https://github.com/France-ioi/AlgoreaBackend/compare/v1.7.0...v1.8.0) - 2023-03-23

- new service: all item breadcrumbs from a `text_id`
- forum: get thread service
- forum: update thread service
- adapt SQL for MySQL 8.0.26 support
- many internal improvements

## [v1.7.0](https://github.com/France-ioi/AlgoreaBackend/compare/v1.6.0...v1.7.0) - 2023-02-22

- new service: get a task token for observation
- make `items.text_id` unique
- new permission "can_request_help_to" (for forum)

## [v1.6.0](https://github.com/France-ioi/AlgoreaBackend/compare/v1.5.0...v1.6.0) - 2023-02-01

- new service: get best answer

## [v1.5.0](https://github.com/France-ioi/AlgoreaBackend/compare/v1.4.0...v1.5.0) - 2023-01-19

- provide 'login' in task token
- add item type in granted permissions view
- add item type in itemBreadcrumbsFromRootsGet
- fix root group service that returned users
- limit item image url to 2048 char
- add type of invisible items in itemChildrenView
- merge item type 'Course' into 'Task'

## [v1.4.0](https://github.com/France-ioi/AlgoreaBackend/compare/v1.3.0...v1.4.0) - 2022-12-09

- fix spec of updatePermissions
- add image_url to get-children service (and other services using the same signature)
- new attribute 'children_layout' for items, update get-item-by-id, create-item and update-item services

## [v1.3.0](https://github.com/France-ioi/AlgoreaBackend/compare/v1.2.0...v1.3.0) - 2022-09-14

- new service: groupParentsView
- implement 'badges' parsing in the user profile
- fix: allow giving permissions to a root activity/skill

## [v1.2.0](https://github.com/France-ioi/AlgoreaBackend/compare/v1.1.0...v1.2.0) - 2022-04-26

- list root content of managed groups in root content services

## [v1.1.0](https://github.com/France-ioi/AlgoreaBackend/compare/v1.0.0...v1.1.0) - 2022-03-29

- fix how bValidated is "computed" in task tokens

## v1.0.0 - 2022-02-15

- initial release, all previous changes can be retrieved through Git history
