Feature: Get recent activity for group_id and item_id
	Background:
		Given the database has the following table 'users':
			| id | login | temp_user | group_self_id | group_owned_id | first_name  | last_name | default_language |
			| 1  | owner | 0         | 21            | 22             | Jean-Michel | Blanquer  | fr               |
			| 2  | user  | 0         | 11            | 12             | John        | Doe       | en               |
			| 3  | jane  | 0         | 31            | 32             | Jane        | Doe       | en               |
		And the database has the following table 'groups_ancestors':
			| id | group_ancestor_id | group_child_id | is_self | version |
			| 75 | 22                | 13             | 0       | 0       |
			| 76 | 13                | 11             | 0       | 0       |
			| 77 | 22                | 11             | 0       | 0       |
			| 78 | 21                | 21             | 1       | 0       |
		And the database has the following table 'users_answers':
			| id | user_id | item_id | attempt_id | name             | type       | state   | lang_prog | submission_date     | score | validated |
			| 2  | 2       | 200     | 101        | My second anwser | Submission | Current | python    | 2017-05-29 06:38:38 | 100   | true      |
			| 1  | 2       | 200     | 100        | My answer        | Submission | Current | python    | 2017-05-29 06:38:38 | 100   | false     |
			| 3  | 2       | 200     | 101        | My third anwser  | Submission | Current | python    | 2017-05-30 06:38:38 | 100   | true      |
			| 4  | 2       | 200     | 101        | My fourth answer | Saved      | Current | python    | 2017-05-30 06:38:38 | 100   | true      |
			| 5  | 2       | 200     | 101        | My fifth answer  | Current    | Current | python    | 2017-05-30 06:38:38 | 100   | true      |
			| 6  | 3       | 200     | 101        | My second anwser | Submission | Current | python    | 2017-05-29 06:38:38 | 100   | true      |
			| 7  | 3       | 200     | 100        | My answer        | Submission | Current | python    | 2017-05-29 06:38:38 | 100   | false     |
			| 8  | 3       | 200     | 101        | My third anwser  | Submission | Current | python    | 2017-05-30 06:38:38 | 100   | true      |
		And the database has the following table 'items':
			| id  | type     | teams_editable | no_score | item_unlocked_id | transparent_folder | version |
			| 200 | Category | false          | false    | 1234,2345        | true               | 0       |
		And the database has the following table 'groups_items':
			| id | group_id | item_id | cached_grayed_access_date | user_created_id | version |
			| 43 | 21       | 200     | 2019-03-22 06:38:38       | 0               | 0       |
		And the database has the following table 'items_ancestors':
			| id | item_ancestor_id | item_child_id | version |
			| 1  | 200              | 200           | 0       |
		And the database has the following table 'items_strings':
			| id | item_id | language_id | title       | image_url                  | subtitle     | description   | edu_comment    | version |
			| 53 | 200     | 1           | Category 1  | http://example.com/my0.jpg | Subtitle 0   | Description 0 | Some comment   | 0       |
			| 63 | 200     | 2           | Catégorie 1 | http://example.com/mf0.jpg | Sous-titre 0 | texte 0       | Un commentaire | 0       |
		And the database has the following table 'languages':
			| id | code |
			| 2  | fr   |

	Scenario: User is an admin of the group and there are visible descendants of the item
		This spec also checks:
	    1) that answers having type!="Submission" are filtered out,
	    2) answers ordering,
	    3) filtering by users groups
		Given I am the user with id "1"
		When I send a GET request to "/groups/13/recent_activity?item_id=200"
		Then the response code should be 200
		And the response body should be, in JSON:
		"""
		[
			{
				"id": "3",
				"item": {
					"id": "200",
					"string": {
						"title": "Catégorie 1"
					},
					"type": "Category"
				},
				"score": 100,
				"submission_date": "2017-05-30T06:38:38Z",
				"user": {
					"first_name": "John",
					"last_name": "Doe",
					"login": "user"
				},
				"validated": true
			},
			{
				"id": "1",
				"item": {
					"id": "200",
					"string": {
						"title": "Catégorie 1"
					},
					"type": "Category"
				},
				"score": 100,
				"submission_date": "2017-05-29T06:38:38Z",
				"user": {
					"first_name": "John",
					"last_name": "Doe",
					"login": "user"
				},
				"validated": false
			},
			{
				"id": "2",
				"item": {
					"id": "200",
					"string": {
						"title": "Catégorie 1"
					},
					"type": "Category"
				},
				"score": 100,
				"submission_date": "2017-05-29T06:38:38Z",
				"user": {
					"first_name": "John",
					"last_name": "Doe",
					"login": "user"
				},
				"validated": true
			}
		]
    """

	Scenario: User is an admin of the group and there are visible descendants of the item; request the first row
		Given I am the user with id "1"
		When I send a GET request to "/groups/13/recent_activity?item_id=200&limit=1"
		Then the response code should be 200
		And the response body should be, in JSON:
		"""
		[
			{
				"id": "3",
				"item": {
					"id": "200",
					"string": {
						"title": "Catégorie 1"
					},
					"type": "Category"
				},
				"score": 100,
				"submission_date": "2017-05-30T06:38:38Z",
				"user": {
					"first_name": "John",
					"last_name": "Doe",
					"login": "user"
				},
				"validated": true
			}
		]
    """

	Scenario: User is an admin of the group and there are visible descendants of the item; request the second and the third rows
		Given I am the user with id "1"
		When I send a GET request to "/groups/13/recent_activity?item_id=200&from.submission_date=2017-05-30T06:38:38Z&from.id=3"
		Then the response code should be 200
		And the response body should be, in JSON:
		"""
		[
			{
				"id": "1",
				"item": {
					"id": "200",
					"string": {
						"title": "Catégorie 1"
					},
					"type": "Category"
				},
				"score": 100,
				"submission_date": "2017-05-29T06:38:38Z",
				"user": {
					"first_name": "John",
					"last_name": "Doe",
					"login": "user"
				},
				"validated": false
			},
			{
				"id": "2",
				"item": {
					"id": "200",
					"string": {
						"title": "Catégorie 1"
					},
					"type": "Category"
				},
				"score": 100,
				"submission_date": "2017-05-29T06:38:38Z",
				"user": {
					"first_name": "John",
					"last_name": "Doe",
					"login": "user"
				},
				"validated": true
			}
	  ]
    """

	Scenario: User is an admin of the group and there are visible descendants of the item; request the third row
		Given I am the user with id "1"
		When I send a GET request to "/groups/13/recent_activity?item_id=200&from.submission_date=2017-05-29T06:38:38Z&from.id=1"
		Then the response code should be 200
		And the response body should be, in JSON:
		"""
		[
			{
				"id": "2",
				"item": {
					"id": "200",
					"string": {
						"title": "Catégorie 1"
					},
					"type": "Category"
				},
				"score": 100,
				"submission_date": "2017-05-29T06:38:38Z",
				"user": {
					"first_name": "John",
					"last_name": "Doe",
					"login": "user"
				},
				"validated": true
			}
	  ]
    """

	Scenario: User is an admin of the group and there are visible descendants of the item; request validated answers only
		Given I am the user with id "1"
		When I send a GET request to "/groups/13/recent_activity?item_id=200&validated=1"
		Then the response code should be 200
		And the response body should be, in JSON:
		"""
		[
			{
				"id": "3",
				"item": {
					"id": "200",
					"string": {
						"title": "Catégorie 1"
					},
					"type": "Category"
				},
				"score": 100,
				"submission_date": "2017-05-30T06:38:38Z",
				"user": {
					"first_name": "John",
					"last_name": "Doe",
					"login": "user"
				},
				"validated": true
			},
			{
				"id": "2",
				"item": {
					"id": "200",
					"string": {
						"title": "Catégorie 1"
					},
					"type": "Category"
				},
				"score": 100,
				"submission_date": "2017-05-29T06:38:38Z",
				"user": {
					"first_name": "John",
					"last_name": "Doe",
					"login": "user"
				},
				"validated": true
			}
		]
    """
