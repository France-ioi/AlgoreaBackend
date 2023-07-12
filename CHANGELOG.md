# Changelog
All notable changes to this project will be documented in this file.

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
