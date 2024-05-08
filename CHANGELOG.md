# Changelog
All notable changes to this project will be documented in this file.

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
