Feature: Display the current progress of users on a subset of items (groupUserProgress)
  Background:
    Given the database has the following table 'users':
      | ID | sLogin | idGroupSelf | idGroupOwned |
      | 1  | owner  | 21          | 22           |
      | 11 | johna  | 51          | 52           |
      | 12 | johnb  | 53          | 54           |
      | 13 | johnc  | 55          | 56           |
      | 14 | johnd  | 57          | 58           |
      | 15 | johne  | 59          | 60           |
      | 16 | janea  | 61          | 62           |
      | 17 | janeb  | 63          | 64           |
      | 18 | janec  | 65          | 66           |
      | 19 | janed  | 67          | 68           |
      | 20 | janee  | 69          | 70           |
    And the database has the following table 'groups':
      | ID | sType     | sName          |
      | 1  | Root      | Root 1         |
      | 3  | Root      | Root 2         |
      | 11 | Class     | Our Class      |
      | 12 | Class     | Other Class    |
      | 13 | Class     | Special Class  |
      | 14 | Team      | Super Team     |
      | 15 | Team      | Our Team       |
      | 16 | Team      | First Team     |
      | 17 | Other     | A custom group |
      | 18 | Club      | Our Club       |
      | 20 | Friends   | My Friends     |
      | 21 | UserSelf  | owner          |
      | 51 | UserSelf  | johna          |
      | 53 | UserSelf  | johnb          |
      | 55 | UserSelf  | johnc          |
      | 57 | UserSelf  | johnd          |
      | 59 | UserSelf  | johne          |
      | 61 | UserSelf  | janea          |
      | 63 | UserSelf  | janeb          |
      | 65 | UserSelf  | janec          |
      | 67 | UserSelf  | janed          |
      | 69 | UserSelf  | janee          |
      | 22 | UserAdmin | owner-admin    |
      | 52 | UserAdmin | johna-admin    |
      | 54 | UserAdmin | johnb-admin    |
      | 56 | UserAdmin | johnc-admin    |
      | 58 | UserAdmin | johnd-admin    |
      | 60 | UserAdmin | johne-admin    |
      | 62 | UserAdmin | janea-admin    |
      | 64 | UserAdmin | janeb-admin    |
      | 66 | UserAdmin | janec-admin    |
      | 68 | UserAdmin | janed-admin    |
      | 70 | UserAdmin | janee-admin    |
    And the database has the following table 'groups_groups':
      | idGroupParent | idGroupChild | sType              |
      | 1             | 11           | direct             |
      | 3             | 13           | direct             |
      | 11            | 14           | direct             |
      | 11            | 16           | direct             |
      | 11            | 17           | direct             |
      | 11            | 18           | direct             |
      | 11            | 59           | requestAccepted    |
      | 13            | 15           | direct             |
      | 13            | 69           | invitationAccepted |
      | 14            | 51           | requestAccepted    |
      | 14            | 53           | requestAccepted    |
      | 14            | 55           | invitationAccepted |
      | 15            | 57           | direct             |
      | 15            | 59           | requestAccepted    |
      | 15            | 61           | invitationAccepted |
      | 15            | 63           | invitationRejected |
      | 15            | 65           | left               |
      | 15            | 67           | invitationSent     |
      | 15            | 69           | requestSent        |
      | 16            | 51           | invitationRefused  |
      | 16            | 53           | requestRefused     |
      | 16            | 55           | removed            |
      | 16            | 63           | direct             |
      | 16            | 65           | requestAccepted    |
      | 16            | 67           | invitationAccepted |
      | 20            | 21           | direct             |
      | 22            | 1            | direct             |
      | 22            | 3            | direct             |
    And the database has the following table 'groups_ancestors':
      | idGroupAncestor | idGroupChild | bIsSelf |
      | 1               | 1            | 1       |
      | 1               | 11           | 0       |
      | 1               | 12           | 0       |
      | 1               | 14           | 0       |
      | 1               | 16           | 0       |
      | 1               | 17           | 0       |
      | 1               | 18           | 0       |
      | 1               | 51           | 0       |
      | 1               | 53           | 0       |
      | 1               | 55           | 0       |
      | 1               | 59           | 0       |
      | 1               | 63           | 0       |
      | 1               | 65           | 0       |
      | 1               | 67           | 0       |
      | 3               | 3            | 1       |
      | 3               | 13           | 0       |
      | 3               | 15           | 0       |
      | 3               | 61           | 0       |
      | 3               | 63           | 0       |
      | 3               | 65           | 0       |
      | 3               | 69           | 0       |
      | 11              | 11           | 1       |
      | 11              | 14           | 0       |
      | 11              | 16           | 0       |
      | 11              | 17           | 0       |
      | 11              | 18           | 0       |
      | 11              | 51           | 0       |
      | 11              | 53           | 0       |
      | 11              | 55           | 0       |
      | 11              | 59           | 0       |
      | 11              | 63           | 0       |
      | 11              | 65           | 0       |
      | 11              | 67           | 0       |
      | 12              | 12           | 1       |
      | 13              | 13           | 1       |
      | 13              | 15           | 0       |
      | 13              | 61           | 0       |
      | 13              | 63           | 0       |
      | 13              | 65           | 0       |
      | 13              | 69           | 0       |
      | 14              | 14           | 1       |
      | 14              | 51           | 0       |
      | 14              | 53           | 0       |
      | 14              | 55           | 0       |
      | 15              | 15           | 1       |
      | 15              | 61           | 0       |
      | 15              | 63           | 0       |
      | 15              | 65           | 0       |
      | 16              | 16           | 1       |
      | 16              | 63           | 0       |
      | 16              | 65           | 0       |
      | 16              | 67           | 0       |
      | 20              | 20           | 1       |
      | 20              | 21           | 0       |
      | 21              | 21           | 1       |
      | 22              | 1            | 0       |
      | 22              | 11           | 0       |
      | 22              | 12           | 0       |
      | 22              | 13           | 0       |
      | 22              | 14           | 0       |
      | 22              | 15           | 0       |
      | 22              | 16           | 0       |
      | 22              | 17           | 0       |
      | 22              | 18           | 0       |
      | 22              | 22           | 1       |
      | 22              | 51           | 0       |
      | 22              | 53           | 0       |
      | 22              | 55           | 0       |
      | 22              | 59           | 0       |
      | 22              | 61           | 0       |
      | 22              | 63           | 0       |
      | 22              | 65           | 0       |
      | 22              | 67           | 0       |
      | 22              | 69           | 0       |
    And the database has the following table 'items':
      | ID  | sType    |
      | 200 | Category |
      | 210 | Chapter  |
      | 211 | Task     |
      | 212 | Task     |
      | 213 | Task     |
      | 214 | Task     |
      | 215 | Task     |
      | 216 | Task     |
      | 217 | Task     |
      | 218 | Task     |
      | 219 | Task     |
      | 220 | Chapter  |
      | 221 | Task     |
      | 222 | Task     |
      | 223 | Task     |
      | 224 | Task     |
      | 225 | Task     |
      | 226 | Task     |
      | 227 | Task     |
      | 228 | Task     |
      | 229 | Task     |
      | 300 | Category |
      | 310 | Chapter  |
      | 311 | Task     |
      | 312 | Task     |
      | 313 | Task     |
      | 314 | Task     |
      | 315 | Task     |
      | 316 | Task     |
      | 317 | Task     |
      | 318 | Task     |
      | 319 | Task     |
      | 400 | Category |
      | 410 | Chapter  |
      | 411 | Task     |
      | 412 | Task     |
      | 413 | Task     |
      | 414 | Task     |
      | 415 | Task     |
      | 416 | Task     |
      | 417 | Task     |
      | 418 | Task     |
      | 419 | Task     |
    And the database has the following table 'items_items':
      | idItemParent | idItemChild |
      | 200          | 210         |
      | 200          | 220         |
      | 210          | 211         |
      | 210          | 212         |
      | 210          | 213         |
      | 210          | 214         |
      | 210          | 215         |
      | 210          | 216         |
      | 210          | 217         |
      | 210          | 218         |
      | 210          | 219         |
      | 220          | 221         |
      | 220          | 222         |
      | 220          | 223         |
      | 220          | 224         |
      | 220          | 225         |
      | 220          | 226         |
      | 220          | 227         |
      | 220          | 228         |
      | 220          | 229         |
      | 300          | 310         |
      | 310          | 311         |
      | 310          | 312         |
      | 310          | 313         |
      | 310          | 314         |
      | 310          | 315         |
      | 310          | 316         |
      | 310          | 317         |
      | 310          | 318         |
      | 310          | 319         |
      | 400          | 410         |
      | 410          | 411         |
      | 410          | 412         |
      | 410          | 413         |
      | 410          | 414         |
      | 410          | 415         |
      | 410          | 416         |
      | 410          | 417         |
      | 410          | 418         |
      | 410          | 419         |
    And the database has the following table 'groups_items':
      | idGroup | idItem | sCachedFullAccessDate | sCachedPartialAccessDate | sCachedGrayedAccessDate |
      | 21      | 211    | null                  | null                     | 2017-05-29T06:38:38Z    |
      | 20      | 212    | null                  | 2017-05-29T06:38:38Z     | null                    |
      | 21      | 213    | 2017-05-29T06:38:38Z  | null                     | null                    |
      | 20      | 214    | null                  | null                     | 2017-05-29T06:38:38Z    |
      | 21      | 215    | null                  | 2017-05-29T06:38:38Z     | null                    |
      | 20      | 216    | null                  | null                     | 2037-05-29T06:38:38Z    |
      | 21      | 217    | null                  | 2037-05-29T06:38:38Z     | null                    |
      | 20      | 218    | 2037-05-29T06:38:38Z  | null                     | null                    |
      | 21      | 219    | null                  | null                     | 2037-05-29T06:38:38Z    |
      | 20      | 221    | null                  | null                     | 2017-05-29T06:38:38Z    |
      | 21      | 222    | null                  | 2017-05-29T06:38:38Z     | null                    |
      | 20      | 223    | 2017-05-29T06:38:38Z  | null                     | null                    |
      | 21      | 224    | null                  | null                     | 2017-05-29T06:38:38Z    |
      | 20      | 225    | null                  | 2017-05-29T06:38:38Z     | null                    |
      | 21      | 226    | null                  | null                     | 2037-05-29T06:38:38Z    |
      | 20      | 227    | null                  | 2037-05-29T06:38:38Z     | null                    |
      | 21      | 228    | 2037-05-29T06:38:38Z  | null                     | null                    |
      | 20      | 229    | null                  | null                     | 2037-05-29T06:38:38Z    |
      | 21      | 311    | null                  | null                     | 2017-05-29T06:38:38Z    |
      | 20      | 312    | null                  | 2017-05-29T06:38:38Z     | null                    |
      | 21      | 313    | 2017-05-29T06:38:38Z  | null                     | null                    |
      | 20      | 314    | null                  | null                     | 2017-05-29T06:38:38Z    |
      | 21      | 315    | null                  | 2017-05-29T06:38:38Z     | null                    |
      | 20      | 316    | null                  | null                     | 2037-05-29T06:38:38Z    |
      | 21      | 317    | null                  | 2037-05-29T06:38:38Z     | null                    |
      | 20      | 318    | 2037-05-29T06:38:38Z  | null                     | null                    |
      | 21      | 319    | null                  | null                     | 2037-05-29T06:38:38Z    |
      | 20      | 411    | null                  | null                     | 2017-05-29T06:38:38Z    |
      | 21      | 412    | null                  | 2017-05-29T06:38:38Z     | null                    |
      | 20      | 413    | 2017-05-29T06:38:38Z  | null                     | null                    |
      | 21      | 414    | null                  | null                     | 2017-05-29T06:38:38Z    |
      | 20      | 415    | null                  | 2017-05-29T06:38:38Z     | null                    |
      | 21      | 416    | null                  | null                     | 2037-05-29T06:38:38Z    |
      | 20      | 417    | null                  | 2037-05-29T06:38:38Z     | null                    |
      | 21      | 418    | 2037-05-29T06:38:38Z  | null                     | null                    |
      | 20      | 419    | null                  | null                     | 2037-05-29T06:38:38Z    |
    And the database has the following table 'groups_attempts':
      | idGroup | idItem | sStartDate           | iScore | sBestAnswerDate      | nbHintsCached | nbSubmissionsAttempts | bValidated | sValidationDate      | sLastActivityDate    |
      | 14      | 211    | 2017-05-29T06:38:38Z | 0      | 2017-05-29T06:38:38Z | 100           | 100                   | 0          | 2017-05-30T06:38:38Z | 2018-05-30T06:38:38Z | # sLastActivityDate for 51, 211 comes from this line (the last activity is made by a team)
      | 14      | 211    | 2017-05-29T06:38:38Z | 40     | 2017-05-29T06:38:38Z | 2             | 3                     | 1          | 2017-05-29T06:38:58Z | null                 | # min(sValidationDate) for 51, 211 comes from this line (from a team)
      | 14      | 211    | 2017-05-29T06:38:38Z | 50     | 2017-05-29T06:38:38Z | 3             | 4                     | 1          | 2017-05-31T06:58:38Z | null                 | # nbHintsCached & nbSubmissionsAttempts for 51, 211 come from this line (the best attempt is made by a team)
      | 14      | 211    | 2017-05-29T06:38:38Z | 50     | 2017-05-30T06:38:38Z | 10            | 20                    | 1          | null                 | null                 |
      | 15      | 211    | 2017-04-29T06:38:38Z | 0      | null                 | 0             | 0                     | 0          | null                 | null                 |
      | 15      | 212    | 2017-03-29T06:38:38Z | 0      | null                 | 0             | 0                     | 0          | null                 | null                 |
      | 16      | 212    | 2018-12-01T00:00:00Z | 10     | 2017-05-30T06:38:38Z | 0             | 0                     | 0          | null                 | 2019-06-01T00:00:00Z | # sStartDate for 67, 212 & 63, 212 comes from this line (the first attempt is started by a team)
      | 67      | 212    | 2019-01-01T00:00:00Z | 20     | 2017-06-30T06:38:38Z | 1             | 2                     | 0          | null                 | 2019-06-01T00:00:00Z | # nbHintsCached & nbSubmissionsAttempts for 67, 212 come from this line (the best attempt is made by a user)
      | 67      | 212    | 2019-01-01T00:00:00Z | 10     | 2017-05-30T06:38:38Z | 6             | 7                     | 0          | null                 | 2019-07-01T00:00:00Z | # sLastActivityDate for 67, 212 comes from this line (the last activity is made by a user)
      | 67      | 213    | 2018-11-01T00:00:00Z | 0      | null                 | 0             | 0                     | 0          | null                 | 2018-11-01T00:00:00Z | # sStartDate for 67, 213 comes from this line (the first attempt is started by a user)
      | 67      | 214    | 2017-05-29T06:38:38Z | 15     | 2017-05-29T06:38:48Z | 10            | 11                    | 1          | 2017-05-29T06:38:48Z | 2017-05-30T06:38:48Z | # min(sValidationDate) for 67, 214 comes from this line (from a user)
      | 14      | 211    | 2017-05-29T06:38:38Z | 0      | null                 | 0             | 0                     | 0          | null                 | null                 |
      | 15      | 211    | 2017-04-29T06:38:38Z | 0      | null                 | 0             | 0                     | 0          | null                 | null                 |
      | 15      | 212    | 2017-03-29T06:38:38Z | 100    | null                 | 0             | 0                     | 1          | null                 | null                 |
      | 14      | 211    | 2017-05-29T06:38:38Z | 0      | null                 | 0             | 0                     | 0          | null                 | null                 |

  Scenario: Get progress of the second and the third users (checks sorting, from.*, and limit)
    Given I am the user with ID "1"
    # here we fixate time_spent even if it depends on NOW()
    And the DB time now is "2019-06-30T20:19:05Z"
    When I send a GET request to "/groups/1/user-progress?parent_item_ids=210&from.name=janec&from.id=65&limit=2"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "group_id": "67",
        "item_id": "211",
        "last_activity_date": null,
        "score": 0,
        "hints_requested": 0,
        "submission_attempts": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "67",
        "hints_requested": 1,
        "item_id": "212",
        "last_activity_date": "2019-07-01T00:00:00Z",
        "score": 20,
        "submission_attempts": 2,
        "time_spent": 18303545,
        "validated": false
      },
      {
        "group_id": "67",
        "hints_requested": 0,
        "item_id": "213",
        "last_activity_date": "2018-11-01T00:00:00Z",
        "score": 0,
        "submission_attempts": 0,
        "time_spent": 20895545,
        "validated": false
      },
      {
        "group_id": "67",
        "hints_requested": 10,
        "item_id": "214",
        "last_activity_date": "2017-05-30T06:38:48Z",
        "score": 15,
        "submission_attempts": 11,
        "time_spent": 10,
        "validated": true
      },
      {
        "group_id": "67",
        "hints_requested": 0,
        "item_id": "215",
        "last_activity_date": null,
        "score": 0,
        "submission_attempts": 0,
        "time_spent": 0,
        "validated": false
      },

      {
        "group_id": "51",
        "item_id": "211",
        "last_activity_date": "2018-05-30T06:38:38Z",
        "score": 50,
        "hints_requested": 3,
        "submission_attempts": 4,
        "time_spent": 20,
        "validated": true
      },
      {
        "group_id": "51",
        "hints_requested": 0,
        "item_id": "212",
        "last_activity_date": null,
        "score": 0,
        "submission_attempts": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "51",
        "hints_requested": 0,
        "item_id": "213",
        "last_activity_date": null,
        "score": 0,
        "submission_attempts": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "51",
        "hints_requested": 0,
        "item_id": "214",
        "last_activity_date": null,
        "score": 0,
        "submission_attempts": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "51",
        "hints_requested": 0,
        "item_id": "215",
        "last_activity_date": null,
        "score": 0,
        "submission_attempts": 0,
        "time_spent": 0,
        "validated": false
      }
    ]
    """


  Scenario: Get progress of the first user for all the visible items (also checks the limit)
    Given I am the user with ID "1"
    # here we fixate time_spent even if it depends on NOW()
    And the DB time now is "2019-06-30T20:19:05Z"
    When I send a GET request to "/groups/1/user-progress?parent_item_ids=210,220,310&limit=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "group_id": "63",
        "item_id": "211",
        "last_activity_date": null,
        "score": 0,
        "hints_requested": 0,
        "submission_attempts": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "63",
        "hints_requested": 0,
        "item_id": "212",
        "last_activity_date": "2019-06-01T00:00:00Z",
        "score": 10,
        "submission_attempts": 0,
        "time_spent": 18303545,
        "validated": false
      },
      {
        "group_id": "63",
        "hints_requested": 0,
        "item_id": "213",
        "last_activity_date": null,
        "score": 0,
        "submission_attempts": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "63",
        "hints_requested": 0,
        "item_id": "214",
        "last_activity_date": null,
        "score": 0,
        "submission_attempts": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "63",
        "hints_requested": 0,
        "item_id": "215",
        "last_activity_date": null,
        "score": 0,
        "submission_attempts": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "63",
        "hints_requested": 0,
        "item_id": "221",
        "last_activity_date": null,
        "score": 0,
        "submission_attempts": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "63",
        "hints_requested": 0,
        "item_id": "222",
        "last_activity_date": null,
        "score": 0,
        "submission_attempts": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "63",
        "hints_requested": 0,
        "item_id": "223",
        "last_activity_date": null,
        "score": 0,
        "submission_attempts": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "63",
        "hints_requested": 0,
        "item_id": "224",
        "last_activity_date": null,
        "score": 0,
        "submission_attempts": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "63",
        "hints_requested": 0,
        "item_id": "225",
        "last_activity_date": null,
        "score": 0,
        "submission_attempts": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "63",
        "hints_requested": 0,
        "item_id": "311",
        "last_activity_date": null,
        "score": 0,
        "submission_attempts": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "63",
        "hints_requested": 0,
        "item_id": "312",
        "last_activity_date": null,
        "score": 0,
        "submission_attempts": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "63",
        "hints_requested": 0,
        "item_id": "313",
        "last_activity_date": null,
        "score": 0,
        "submission_attempts": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "63",
        "hints_requested": 0,
        "item_id": "314",
        "last_activity_date": null,
        "score": 0,
        "submission_attempts": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "63",
        "hints_requested": 0,
        "item_id": "315",
        "last_activity_date": null,
        "score": 0,
        "submission_attempts": 0,
        "time_spent": 0,
        "validated": false
      }
    ]
    """

  Scenario: No users
    Given I am the user with ID "1"
    When I send a GET request to "/groups/17/user-progress?parent_item_ids=210,220,310"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
    ]
    """

  Scenario: No visible items
    Given I am the user with ID "1"
    When I send a GET request to "/groups/1/user-progress?parent_item_ids=1010"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
    ]
    """
