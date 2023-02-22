# Changelog
All notable changes to this project will be documented in this file.

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
