Feature: Export the current progress of a group with answers on a subset of items as a ZIP file (groupGroupProgressWithAnswersZIP) - robustness
  Scenario: User is not able to watch group members
    Given I am @John
    And there is a group @Classroom
    And there are the following items:
      | item  | type |
      | @Item | Task |
    And I am a manager of the group @Classroom
    When I send a GET request to "/groups/@Classroom/group-progress-with-answers-zip?parent_item_ids=@Item"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: Group id is incorrect
    Given I am @John
    And there are the following items:
      | item  | type |
      | @Item | Task |
    When I send a GET request to "/groups/abc/group-progress-with-answers-zip?parent_item_ids=@Item"
    Then the response code should be 400
    And the response error message should contain "Wrong value for group_id (should be int64)"

  Scenario: parent_item_ids is incorrect
    Given I am @John
    And there is a group @Classroom
    And I am a manager of the group @Classroom and can watch for submissions from the group and its descendants
    And there are the following items:
      | item  | type |
      | @Item | Task |
    When I send a GET request to "/groups/@Classroom/group-progress-with-answers-zip?parent_item_ids=abc,@Item"
    Then the response code should be 400
    And the response error message should contain "Unable to parse one of the integers given as query args (value: 'abc', param: 'parent_item_ids')"

  Scenario: Not enough permissions to watch answers on parent_item_ids
    Given I am @John
    And there is a group @Classroom
    And I am a manager of the group @Classroom and can watch for submissions from the group and its descendants
    And there are the following items:
      | item                | type |
      | @ItemCanWatchAnswer | Task |
      | @ItemCanWatchResult | Task |
    And there are the following item permissions:
      | item                | group | can_watch |
      | @ItemCanWatchAnswer | @John | answer    |
      | @ItemCanWatchResult | @John | result    |
    When I send a GET request to "/groups/@Classroom/group-progress-with-answers-zip?parent_item_ids=@ItemCanWatchAnswer,@ItemCanWatchResult"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: User not found
    Given I am the user with id "404"
    And there is a group @Classroom
    And there are the following items:
      | item  | type |
      | @Item | Task |
    When I send a GET request to "/groups/@Classroom/group-progress-with-answers-zip?parent_item_ids=@Item"
    Then the response code should be 401
    And the response error message should contain "Invalid access token"


  Scenario: The number of items exceeds the limit
    Given I am the user with id "21"
    And the database has the following table "groups":
      | id | type  | name      |
      | 1  | Base  | Root 1    |
      | 11 | Class | Our Class |
    And the database has the following users:
      | group_id | login | default_language |
      | 21       | owner | en               |
    And the database has the following table "group_managers":
      | group_id | manager_id | can_watch_members |
      | 1        | 21         | true              |
    And the database has the following table "groups_groups":
      | parent_group_id | child_group_id | is_team_membership |
      | 1               | 11             | 0                  |
    And the groups ancestors are computed
    And the database has the following table "items":
      | id | type | default_language_tag |
      | 6000 | Chapter | fr |
      | 6001 | Task | fr |
      | 6002 | Task | fr |
      | 6003 | Task | fr |
      | 6004 | Task | fr |
      | 6005 | Task | fr |
      | 6006 | Task | fr |
      | 6007 | Task | fr |
      | 6008 | Task | fr |
      | 6009 | Task | fr |
      | 6010 | Task | fr |
      | 6011 | Task | fr |
      | 6012 | Task | fr |
      | 6013 | Task | fr |
      | 6014 | Task | fr |
      | 6015 | Task | fr |
      | 6016 | Task | fr |
      | 6017 | Task | fr |
      | 6018 | Task | fr |
      | 6019 | Task | fr |
      | 6020 | Task | fr |
      | 6021 | Task | fr |
      | 6022 | Task | fr |
      | 6023 | Task | fr |
      | 6024 | Task | fr |
      | 6025 | Task | fr |
      | 6026 | Task | fr |
      | 6027 | Task | fr |
      | 6028 | Task | fr |
      | 6029 | Task | fr |
      | 6030 | Task | fr |
      | 6031 | Task | fr |
      | 6032 | Task | fr |
      | 6033 | Task | fr |
      | 6034 | Task | fr |
      | 6035 | Task | fr |
      | 6036 | Task | fr |
      | 6037 | Task | fr |
      | 6038 | Task | fr |
      | 6039 | Task | fr |
      | 6040 | Task | fr |
      | 6041 | Task | fr |
      | 6042 | Task | fr |
      | 6043 | Task | fr |
      | 6044 | Task | fr |
      | 6045 | Task | fr |
      | 6046 | Task | fr |
      | 6047 | Task | fr |
      | 6048 | Task | fr |
      | 6049 | Task | fr |
      | 6050 | Task | fr |
      | 6051 | Task | fr |
      | 6052 | Task | fr |
      | 6053 | Task | fr |
      | 6054 | Task | fr |
      | 6055 | Task | fr |
      | 6056 | Task | fr |
      | 6057 | Task | fr |
      | 6058 | Task | fr |
      | 6059 | Task | fr |
      | 6060 | Task | fr |
      | 6061 | Task | fr |
      | 6062 | Task | fr |
      | 6063 | Task | fr |
      | 6064 | Task | fr |
      | 6065 | Task | fr |
      | 6066 | Task | fr |
      | 6067 | Task | fr |
      | 6068 | Task | fr |
      | 6069 | Task | fr |
      | 6070 | Task | fr |
      | 6071 | Task | fr |
      | 6072 | Task | fr |
      | 6073 | Task | fr |
      | 6074 | Task | fr |
      | 6075 | Task | fr |
      | 6076 | Task | fr |
      | 6077 | Task | fr |
      | 6078 | Task | fr |
      | 6079 | Task | fr |
      | 6080 | Task | fr |
      | 6081 | Task | fr |
      | 6082 | Task | fr |
      | 6083 | Task | fr |
      | 6084 | Task | fr |
      | 6085 | Task | fr |
      | 6086 | Task | fr |
      | 6087 | Task | fr |
      | 6088 | Task | fr |
      | 6089 | Task | fr |
      | 6090 | Task | fr |
      | 6091 | Task | fr |
      | 6092 | Task | fr |
      | 6093 | Task | fr |
      | 6094 | Task | fr |
      | 6095 | Task | fr |
      | 6096 | Task | fr |
      | 6097 | Task | fr |
      | 6098 | Task | fr |
      | 6099 | Task | fr |
      | 6100 | Task | fr |
    And the database has the following table "items_items":
      | parent_item_id | child_item_id | child_order |
      | 6000 | 6001 | 0 |
      | 6000 | 6002 | 1 |
      | 6000 | 6003 | 2 |
      | 6000 | 6004 | 3 |
      | 6000 | 6005 | 4 |
      | 6000 | 6006 | 5 |
      | 6000 | 6007 | 6 |
      | 6000 | 6008 | 7 |
      | 6000 | 6009 | 8 |
      | 6000 | 6010 | 9 |
      | 6000 | 6011 | 10 |
      | 6000 | 6012 | 11 |
      | 6000 | 6013 | 12 |
      | 6000 | 6014 | 13 |
      | 6000 | 6015 | 14 |
      | 6000 | 6016 | 15 |
      | 6000 | 6017 | 16 |
      | 6000 | 6018 | 17 |
      | 6000 | 6019 | 18 |
      | 6000 | 6020 | 19 |
      | 6000 | 6021 | 20 |
      | 6000 | 6022 | 21 |
      | 6000 | 6023 | 22 |
      | 6000 | 6024 | 23 |
      | 6000 | 6025 | 24 |
      | 6000 | 6026 | 25 |
      | 6000 | 6027 | 26 |
      | 6000 | 6028 | 27 |
      | 6000 | 6029 | 28 |
      | 6000 | 6030 | 29 |
      | 6000 | 6031 | 30 |
      | 6000 | 6032 | 31 |
      | 6000 | 6033 | 32 |
      | 6000 | 6034 | 33 |
      | 6000 | 6035 | 34 |
      | 6000 | 6036 | 35 |
      | 6000 | 6037 | 36 |
      | 6000 | 6038 | 37 |
      | 6000 | 6039 | 38 |
      | 6000 | 6040 | 39 |
      | 6000 | 6041 | 40 |
      | 6000 | 6042 | 41 |
      | 6000 | 6043 | 42 |
      | 6000 | 6044 | 43 |
      | 6000 | 6045 | 44 |
      | 6000 | 6046 | 45 |
      | 6000 | 6047 | 46 |
      | 6000 | 6048 | 47 |
      | 6000 | 6049 | 48 |
      | 6000 | 6050 | 49 |
      | 6000 | 6051 | 50 |
      | 6000 | 6052 | 51 |
      | 6000 | 6053 | 52 |
      | 6000 | 6054 | 53 |
      | 6000 | 6055 | 54 |
      | 6000 | 6056 | 55 |
      | 6000 | 6057 | 56 |
      | 6000 | 6058 | 57 |
      | 6000 | 6059 | 58 |
      | 6000 | 6060 | 59 |
      | 6000 | 6061 | 60 |
      | 6000 | 6062 | 61 |
      | 6000 | 6063 | 62 |
      | 6000 | 6064 | 63 |
      | 6000 | 6065 | 64 |
      | 6000 | 6066 | 65 |
      | 6000 | 6067 | 66 |
      | 6000 | 6068 | 67 |
      | 6000 | 6069 | 68 |
      | 6000 | 6070 | 69 |
      | 6000 | 6071 | 70 |
      | 6000 | 6072 | 71 |
      | 6000 | 6073 | 72 |
      | 6000 | 6074 | 73 |
      | 6000 | 6075 | 74 |
      | 6000 | 6076 | 75 |
      | 6000 | 6077 | 76 |
      | 6000 | 6078 | 77 |
      | 6000 | 6079 | 78 |
      | 6000 | 6080 | 79 |
      | 6000 | 6081 | 80 |
      | 6000 | 6082 | 81 |
      | 6000 | 6083 | 82 |
      | 6000 | 6084 | 83 |
      | 6000 | 6085 | 84 |
      | 6000 | 6086 | 85 |
      | 6000 | 6087 | 86 |
      | 6000 | 6088 | 87 |
      | 6000 | 6089 | 88 |
      | 6000 | 6090 | 89 |
      | 6000 | 6091 | 90 |
      | 6000 | 6092 | 91 |
      | 6000 | 6093 | 92 |
      | 6000 | 6094 | 93 |
      | 6000 | 6095 | 94 |
      | 6000 | 6096 | 95 |
      | 6000 | 6097 | 96 |
      | 6000 | 6098 | 97 |
      | 6000 | 6099 | 98 |
      | 6000 | 6100 | 99 |
    And the database has the following table "permissions_generated":
      | group_id | item_id | can_view_generated | can_watch_generated |
      | 21 | 6000 | info | answer |
      | 21 | 6001 | info | none |
      | 21 | 6002 | info | none |
      | 21 | 6003 | info | none |
      | 21 | 6004 | info | none |
      | 21 | 6005 | info | none |
      | 21 | 6006 | info | none |
      | 21 | 6007 | info | none |
      | 21 | 6008 | info | none |
      | 21 | 6009 | info | none |
      | 21 | 6010 | info | none |
      | 21 | 6011 | info | none |
      | 21 | 6012 | info | none |
      | 21 | 6013 | info | none |
      | 21 | 6014 | info | none |
      | 21 | 6015 | info | none |
      | 21 | 6016 | info | none |
      | 21 | 6017 | info | none |
      | 21 | 6018 | info | none |
      | 21 | 6019 | info | none |
      | 21 | 6020 | info | none |
      | 21 | 6021 | info | none |
      | 21 | 6022 | info | none |
      | 21 | 6023 | info | none |
      | 21 | 6024 | info | none |
      | 21 | 6025 | info | none |
      | 21 | 6026 | info | none |
      | 21 | 6027 | info | none |
      | 21 | 6028 | info | none |
      | 21 | 6029 | info | none |
      | 21 | 6030 | info | none |
      | 21 | 6031 | info | none |
      | 21 | 6032 | info | none |
      | 21 | 6033 | info | none |
      | 21 | 6034 | info | none |
      | 21 | 6035 | info | none |
      | 21 | 6036 | info | none |
      | 21 | 6037 | info | none |
      | 21 | 6038 | info | none |
      | 21 | 6039 | info | none |
      | 21 | 6040 | info | none |
      | 21 | 6041 | info | none |
      | 21 | 6042 | info | none |
      | 21 | 6043 | info | none |
      | 21 | 6044 | info | none |
      | 21 | 6045 | info | none |
      | 21 | 6046 | info | none |
      | 21 | 6047 | info | none |
      | 21 | 6048 | info | none |
      | 21 | 6049 | info | none |
      | 21 | 6050 | info | none |
      | 21 | 6051 | info | none |
      | 21 | 6052 | info | none |
      | 21 | 6053 | info | none |
      | 21 | 6054 | info | none |
      | 21 | 6055 | info | none |
      | 21 | 6056 | info | none |
      | 21 | 6057 | info | none |
      | 21 | 6058 | info | none |
      | 21 | 6059 | info | none |
      | 21 | 6060 | info | none |
      | 21 | 6061 | info | none |
      | 21 | 6062 | info | none |
      | 21 | 6063 | info | none |
      | 21 | 6064 | info | none |
      | 21 | 6065 | info | none |
      | 21 | 6066 | info | none |
      | 21 | 6067 | info | none |
      | 21 | 6068 | info | none |
      | 21 | 6069 | info | none |
      | 21 | 6070 | info | none |
      | 21 | 6071 | info | none |
      | 21 | 6072 | info | none |
      | 21 | 6073 | info | none |
      | 21 | 6074 | info | none |
      | 21 | 6075 | info | none |
      | 21 | 6076 | info | none |
      | 21 | 6077 | info | none |
      | 21 | 6078 | info | none |
      | 21 | 6079 | info | none |
      | 21 | 6080 | info | none |
      | 21 | 6081 | info | none |
      | 21 | 6082 | info | none |
      | 21 | 6083 | info | none |
      | 21 | 6084 | info | none |
      | 21 | 6085 | info | none |
      | 21 | 6086 | info | none |
      | 21 | 6087 | info | none |
      | 21 | 6088 | info | none |
      | 21 | 6089 | info | none |
      | 21 | 6090 | info | none |
      | 21 | 6091 | info | none |
      | 21 | 6092 | info | none |
      | 21 | 6093 | info | none |
      | 21 | 6094 | info | none |
      | 21 | 6095 | info | none |
      | 21 | 6096 | info | none |
      | 21 | 6097 | info | none |
      | 21 | 6098 | info | none |
      | 21 | 6099 | info | none |
      | 21 | 6100 | info | none |
    When I send a GET request to "/groups/11/group-progress-with-answers-zip?parent_item_ids=6000"
    Then the response code should be 400
    And the response error message should contain "The number of items exceeds the limit (100)"


  Scenario: The number of users exceeds the limit
    Given I am the user with id "21"
    And the database has the following table "groups":
      | id | type  | name      |
      | 1  | Base  | Root 1    |
      | 11 | Class | Our Class |
      | 21 | User  | owner     |
      | 7000 | User | user000 |
      | 7001 | User | user001 |
      | 7002 | User | user002 |
      | 7003 | User | user003 |
      | 7004 | User | user004 |
      | 7005 | User | user005 |
      | 7006 | User | user006 |
      | 7007 | User | user007 |
      | 7008 | User | user008 |
      | 7009 | User | user009 |
      | 7010 | User | user010 |
      | 7011 | User | user011 |
      | 7012 | User | user012 |
      | 7013 | User | user013 |
      | 7014 | User | user014 |
      | 7015 | User | user015 |
      | 7016 | User | user016 |
      | 7017 | User | user017 |
      | 7018 | User | user018 |
      | 7019 | User | user019 |
      | 7020 | User | user020 |
      | 7021 | User | user021 |
      | 7022 | User | user022 |
      | 7023 | User | user023 |
      | 7024 | User | user024 |
      | 7025 | User | user025 |
      | 7026 | User | user026 |
      | 7027 | User | user027 |
      | 7028 | User | user028 |
      | 7029 | User | user029 |
      | 7030 | User | user030 |
      | 7031 | User | user031 |
      | 7032 | User | user032 |
      | 7033 | User | user033 |
      | 7034 | User | user034 |
      | 7035 | User | user035 |
      | 7036 | User | user036 |
      | 7037 | User | user037 |
      | 7038 | User | user038 |
      | 7039 | User | user039 |
      | 7040 | User | user040 |
      | 7041 | User | user041 |
      | 7042 | User | user042 |
      | 7043 | User | user043 |
      | 7044 | User | user044 |
      | 7045 | User | user045 |
      | 7046 | User | user046 |
      | 7047 | User | user047 |
      | 7048 | User | user048 |
      | 7049 | User | user049 |
      | 7050 | User | user050 |
      | 7051 | User | user051 |
      | 7052 | User | user052 |
      | 7053 | User | user053 |
      | 7054 | User | user054 |
      | 7055 | User | user055 |
      | 7056 | User | user056 |
      | 7057 | User | user057 |
      | 7058 | User | user058 |
      | 7059 | User | user059 |
      | 7060 | User | user060 |
      | 7061 | User | user061 |
      | 7062 | User | user062 |
      | 7063 | User | user063 |
      | 7064 | User | user064 |
      | 7065 | User | user065 |
      | 7066 | User | user066 |
      | 7067 | User | user067 |
      | 7068 | User | user068 |
      | 7069 | User | user069 |
      | 7070 | User | user070 |
      | 7071 | User | user071 |
      | 7072 | User | user072 |
      | 7073 | User | user073 |
      | 7074 | User | user074 |
      | 7075 | User | user075 |
      | 7076 | User | user076 |
      | 7077 | User | user077 |
      | 7078 | User | user078 |
      | 7079 | User | user079 |
      | 7080 | User | user080 |
      | 7081 | User | user081 |
      | 7082 | User | user082 |
      | 7083 | User | user083 |
      | 7084 | User | user084 |
      | 7085 | User | user085 |
      | 7086 | User | user086 |
      | 7087 | User | user087 |
      | 7088 | User | user088 |
      | 7089 | User | user089 |
      | 7090 | User | user090 |
      | 7091 | User | user091 |
      | 7092 | User | user092 |
      | 7093 | User | user093 |
      | 7094 | User | user094 |
      | 7095 | User | user095 |
      | 7096 | User | user096 |
      | 7097 | User | user097 |
      | 7098 | User | user098 |
      | 7099 | User | user099 |
      | 7100 | User | user100 |
    And the database has the following users:
      | group_id | login | default_language |
      | 21       | owner | en               |
      | 7000 | user000 | fr |
      | 7001 | user001 | fr |
      | 7002 | user002 | fr |
      | 7003 | user003 | fr |
      | 7004 | user004 | fr |
      | 7005 | user005 | fr |
      | 7006 | user006 | fr |
      | 7007 | user007 | fr |
      | 7008 | user008 | fr |
      | 7009 | user009 | fr |
      | 7010 | user010 | fr |
      | 7011 | user011 | fr |
      | 7012 | user012 | fr |
      | 7013 | user013 | fr |
      | 7014 | user014 | fr |
      | 7015 | user015 | fr |
      | 7016 | user016 | fr |
      | 7017 | user017 | fr |
      | 7018 | user018 | fr |
      | 7019 | user019 | fr |
      | 7020 | user020 | fr |
      | 7021 | user021 | fr |
      | 7022 | user022 | fr |
      | 7023 | user023 | fr |
      | 7024 | user024 | fr |
      | 7025 | user025 | fr |
      | 7026 | user026 | fr |
      | 7027 | user027 | fr |
      | 7028 | user028 | fr |
      | 7029 | user029 | fr |
      | 7030 | user030 | fr |
      | 7031 | user031 | fr |
      | 7032 | user032 | fr |
      | 7033 | user033 | fr |
      | 7034 | user034 | fr |
      | 7035 | user035 | fr |
      | 7036 | user036 | fr |
      | 7037 | user037 | fr |
      | 7038 | user038 | fr |
      | 7039 | user039 | fr |
      | 7040 | user040 | fr |
      | 7041 | user041 | fr |
      | 7042 | user042 | fr |
      | 7043 | user043 | fr |
      | 7044 | user044 | fr |
      | 7045 | user045 | fr |
      | 7046 | user046 | fr |
      | 7047 | user047 | fr |
      | 7048 | user048 | fr |
      | 7049 | user049 | fr |
      | 7050 | user050 | fr |
      | 7051 | user051 | fr |
      | 7052 | user052 | fr |
      | 7053 | user053 | fr |
      | 7054 | user054 | fr |
      | 7055 | user055 | fr |
      | 7056 | user056 | fr |
      | 7057 | user057 | fr |
      | 7058 | user058 | fr |
      | 7059 | user059 | fr |
      | 7060 | user060 | fr |
      | 7061 | user061 | fr |
      | 7062 | user062 | fr |
      | 7063 | user063 | fr |
      | 7064 | user064 | fr |
      | 7065 | user065 | fr |
      | 7066 | user066 | fr |
      | 7067 | user067 | fr |
      | 7068 | user068 | fr |
      | 7069 | user069 | fr |
      | 7070 | user070 | fr |
      | 7071 | user071 | fr |
      | 7072 | user072 | fr |
      | 7073 | user073 | fr |
      | 7074 | user074 | fr |
      | 7075 | user075 | fr |
      | 7076 | user076 | fr |
      | 7077 | user077 | fr |
      | 7078 | user078 | fr |
      | 7079 | user079 | fr |
      | 7080 | user080 | fr |
      | 7081 | user081 | fr |
      | 7082 | user082 | fr |
      | 7083 | user083 | fr |
      | 7084 | user084 | fr |
      | 7085 | user085 | fr |
      | 7086 | user086 | fr |
      | 7087 | user087 | fr |
      | 7088 | user088 | fr |
      | 7089 | user089 | fr |
      | 7090 | user090 | fr |
      | 7091 | user091 | fr |
      | 7092 | user092 | fr |
      | 7093 | user093 | fr |
      | 7094 | user094 | fr |
      | 7095 | user095 | fr |
      | 7096 | user096 | fr |
      | 7097 | user097 | fr |
      | 7098 | user098 | fr |
      | 7099 | user099 | fr |
      | 7100 | user100 | fr |
    And the database has the following table "group_managers":
      | group_id | manager_id | can_watch_members |
      | 1        | 21         | true              |
    And the database has the following table "groups_groups":
      | parent_group_id | child_group_id | is_team_membership |
      | 1 | 11 | 0 |
      | 11 | 7000 | 0 |
      | 11 | 7001 | 0 |
      | 11 | 7002 | 0 |
      | 11 | 7003 | 0 |
      | 11 | 7004 | 0 |
      | 11 | 7005 | 0 |
      | 11 | 7006 | 0 |
      | 11 | 7007 | 0 |
      | 11 | 7008 | 0 |
      | 11 | 7009 | 0 |
      | 11 | 7010 | 0 |
      | 11 | 7011 | 0 |
      | 11 | 7012 | 0 |
      | 11 | 7013 | 0 |
      | 11 | 7014 | 0 |
      | 11 | 7015 | 0 |
      | 11 | 7016 | 0 |
      | 11 | 7017 | 0 |
      | 11 | 7018 | 0 |
      | 11 | 7019 | 0 |
      | 11 | 7020 | 0 |
      | 11 | 7021 | 0 |
      | 11 | 7022 | 0 |
      | 11 | 7023 | 0 |
      | 11 | 7024 | 0 |
      | 11 | 7025 | 0 |
      | 11 | 7026 | 0 |
      | 11 | 7027 | 0 |
      | 11 | 7028 | 0 |
      | 11 | 7029 | 0 |
      | 11 | 7030 | 0 |
      | 11 | 7031 | 0 |
      | 11 | 7032 | 0 |
      | 11 | 7033 | 0 |
      | 11 | 7034 | 0 |
      | 11 | 7035 | 0 |
      | 11 | 7036 | 0 |
      | 11 | 7037 | 0 |
      | 11 | 7038 | 0 |
      | 11 | 7039 | 0 |
      | 11 | 7040 | 0 |
      | 11 | 7041 | 0 |
      | 11 | 7042 | 0 |
      | 11 | 7043 | 0 |
      | 11 | 7044 | 0 |
      | 11 | 7045 | 0 |
      | 11 | 7046 | 0 |
      | 11 | 7047 | 0 |
      | 11 | 7048 | 0 |
      | 11 | 7049 | 0 |
      | 11 | 7050 | 0 |
      | 11 | 7051 | 0 |
      | 11 | 7052 | 0 |
      | 11 | 7053 | 0 |
      | 11 | 7054 | 0 |
      | 11 | 7055 | 0 |
      | 11 | 7056 | 0 |
      | 11 | 7057 | 0 |
      | 11 | 7058 | 0 |
      | 11 | 7059 | 0 |
      | 11 | 7060 | 0 |
      | 11 | 7061 | 0 |
      | 11 | 7062 | 0 |
      | 11 | 7063 | 0 |
      | 11 | 7064 | 0 |
      | 11 | 7065 | 0 |
      | 11 | 7066 | 0 |
      | 11 | 7067 | 0 |
      | 11 | 7068 | 0 |
      | 11 | 7069 | 0 |
      | 11 | 7070 | 0 |
      | 11 | 7071 | 0 |
      | 11 | 7072 | 0 |
      | 11 | 7073 | 0 |
      | 11 | 7074 | 0 |
      | 11 | 7075 | 0 |
      | 11 | 7076 | 0 |
      | 11 | 7077 | 0 |
      | 11 | 7078 | 0 |
      | 11 | 7079 | 0 |
      | 11 | 7080 | 0 |
      | 11 | 7081 | 0 |
      | 11 | 7082 | 0 |
      | 11 | 7083 | 0 |
      | 11 | 7084 | 0 |
      | 11 | 7085 | 0 |
      | 11 | 7086 | 0 |
      | 11 | 7087 | 0 |
      | 11 | 7088 | 0 |
      | 11 | 7089 | 0 |
      | 11 | 7090 | 0 |
      | 11 | 7091 | 0 |
      | 11 | 7092 | 0 |
      | 11 | 7093 | 0 |
      | 11 | 7094 | 0 |
      | 11 | 7095 | 0 |
      | 11 | 7096 | 0 |
      | 11 | 7097 | 0 |
      | 11 | 7098 | 0 |
      | 11 | 7099 | 0 |
      | 11 | 7100 | 0 |
    And the groups ancestors are computed
    And the database has the following table "items":
      | id  | type | default_language_tag |
      | 800 | Task | fr                   |
    And the database has the following table "permissions_generated":
      | group_id | item_id | can_view_generated | can_watch_generated |
      | 21       | 800     | info               | answer              |
    When I send a GET request to "/groups/11/group-progress-with-answers-zip?parent_item_ids=800"
    Then the response code should be 400
    And the response error message should contain "The number of users exceeds the limit (100)"
