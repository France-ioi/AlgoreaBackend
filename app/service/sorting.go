package service

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/France-ioi/AlgoreaBackend/app/database"
)

// FieldSortingParams represents sorting parameters for one field
type FieldSortingParams struct {
	// ColumnName is a DB column name (may contain a table name as a prefix, e.g. "groups.ID")
	ColumnName string
	// FieldType is one of "int64", "bool", "string"
	FieldType string
}

// ApplySorting applies ordering and paging according to given accepted fields and sorting rules
// taking into the account the URL parameters 'from.*'
func ApplySorting(r *http.Request, query *database.DB, acceptedFields map[string]*FieldSortingParams,
	defaultRules string) (*database.DB, APIError) {
	sortingRules, acceptedFields := prepareSoringRulesAndAcceptedFields(r, acceptedFields, defaultRules)

	usedFields, fieldsDirections, err := parseSortingRules(sortingRules, acceptedFields)
	if err != nil {
		return nil, ErrInvalidRequest(err)
	}
	query = applyOrder(query, usedFields, acceptedFields, fieldsDirections)

	fromValues, err := parsePagingParameters(r, usedFields, acceptedFields)
	if err != nil {
		return nil, ErrInvalidRequest(err)
	}

	query = applyPagingConditions(query, usedFields, fieldsDirections, acceptedFields, fromValues)
	return query, NoError
}

func prepareSoringRulesAndAcceptedFields(r *http.Request, acceptedFields map[string]*FieldSortingParams,
	defaultRules string) (string, map[string]*FieldSortingParams) {
	newAcceptedFields := make(map[string]*FieldSortingParams, len(acceptedFields)+1)
	for field, params := range acceptedFields {
		newAcceptedFields[field] = params
	}
	if _, ok := newAcceptedFields["id"]; !ok {
		newAcceptedFields["id"] = &FieldSortingParams{ColumnName: "ID", FieldType: "int64"}
	}
	var sort string
	urlQuery := r.URL.Query()
	if len(urlQuery["sort"]) > 0 {
		sort = urlQuery["sort"][0]
	} else {
		sort = defaultRules
	}
	return sort, newAcceptedFields
}

func parseSortingRules(sortingRules string,
	acceptedFields map[string]*FieldSortingParams) (usedFields []string, fieldsDirections map[string]int, err error) {
	sortStatements := strings.Split(sortingRules, ",")
	usedFields = make([]string, 0, len(sortStatements)+1)
	fieldsDirections = make(map[string]int, len(sortStatements)+1)
	for _, sortStatement := range sortStatements {
		fieldName, direction := getFieldNameAndDirectionFromSortStatement(sortStatement)
		if fieldsDirections[fieldName] != 0 {
			return nil, nil, fmt.Errorf("a field cannot be a sorting parameter more than once: %q", fieldName)
		}
		if _, ok := acceptedFields[fieldName]; !ok {
			return nil, nil, fmt.Errorf("unknown field in sorting parameters: %q", fieldName)
		}
		fieldsDirections[fieldName] = direction
		usedFields = append(usedFields, fieldName)
	}
	if fieldsDirections["id"] == 0 {
		fieldsDirections["id"] = 1
		usedFields = append(usedFields, "id")
	}
	return
}

func getFieldNameAndDirectionFromSortStatement(sortStatement string) (string, int) {
	var direction int
	if len(sortStatement) > 0 && sortStatement[0] == '-' {
		sortStatement = sortStatement[1:]
		direction = -1
	} else {
		direction = 1
	}
	fieldName := sortStatement
	return fieldName, direction
}

func applyOrder(query *database.DB, usedFields []string, acceptedFields map[string]*FieldSortingParams,
	fieldsDirections map[string]int) *database.DB {
	usedFieldsNumber := len(usedFields)
	orderStrings := make([]string, 0, usedFieldsNumber)
	for _, field := range usedFields {
		var directionString string
		if fieldsDirections[field] == 1 {
			directionString = "ASC"
		} else {
			directionString = "DESC"
		}
		orderStrings = append(orderStrings, acceptedFields[field].ColumnName+" "+directionString)
	}
	if len(orderStrings) > 0 {
		query = query.Order(strings.Join(orderStrings, ", "))
	}
	return query
}

func parsePagingParameters(r *http.Request, usedFields []string,
	acceptedFields map[string]*FieldSortingParams) ([]interface{}, error) {
	fromValueSkipped := false
	fromValueAccepted := false
	fromValues := make([]interface{}, 0, len(usedFields))
	for _, fieldName := range usedFields {
		var value interface{}
		fromFieldName := "from." + fieldName
		if len(r.URL.Query()[fromFieldName]) > 0 {
			var err error
			value, err = getFromValueForField(r, fieldName, acceptedFields)
			if err != nil {
				return nil, err
			}
			fromValueAccepted = true
		} else {
			fromValueSkipped = true
			continue
		}
		fromValues = append(fromValues, value)
	}
	if fromValueAccepted && fromValueSkipped {
		fromParameters := strings.Join(usedFields, ", from.")
		return nil, fmt.Errorf("all 'from' parameters (from.%s) or none of them must be present", fromParameters)
	}
	return fromValues, nil
}

func getFromValueForField(r *http.Request, fieldName string,
	acceptedFields map[string]*FieldSortingParams) (interface{}, error) {
	fromFieldName := "from." + fieldName
	switch acceptedFields[fieldName].FieldType {
	case "string":
		return r.URL.Query()[fromFieldName][0], nil
	case "int64":
		return ResolveURLQueryGetInt64Field(r, fromFieldName)
	case "bool":
		return ResolveURLQueryGetBoolField(r, fromFieldName)
	default:
		return nil, fmt.Errorf("unsupported type '%s' for field '%s", acceptedFields[fieldName].FieldType, fieldName)
	}
}

func applyPagingConditions(query *database.DB, usedFields []string, fieldsDirections map[string]int,
	acceptedFields map[string]*FieldSortingParams, fromValues []interface{}) *database.DB {
	if len(fromValues) == 0 {
		return query
	}
	usedFieldsNumber := len(usedFields)
	conditions := make([]string, 0, usedFieldsNumber)
	queryValues := make([]interface{}, 0, (usedFieldsNumber+1)*usedFieldsNumber/2)
	queryValuesPart := make([]interface{}, 0, usedFieldsNumber)
	conditionPrefix := ""

	// Here we're constructing an expression like this one:
	// (col1 > val1) OR (col1 = val1 AND col2 > val2) OR (col1 = val1 AND col2 = val2 AND col3 > val3) OR ...
	for index, fieldName := range usedFields {
		if len(conditionPrefix) > 0 {
			conditionPrefix += " AND "
		}
		conditionSign := ">"
		if fieldsDirections[fieldName] == -1 {
			conditionSign = "<"
		}
		conditions = append(conditions,
			fmt.Sprintf("(%s%s %s ?)", conditionPrefix, acceptedFields[fieldName].ColumnName, conditionSign))
		conditionPrefix = fmt.Sprintf("%s%s = ?", conditionPrefix, acceptedFields[fieldName].ColumnName)

		queryValuesPart = append(queryValuesPart, fromValues[index])
		queryValues = append(queryValues, queryValuesPart...)
	}
	if len(conditions) > 0 {
		query = query.Where(strings.Join(conditions, " OR "), queryValues...)
	}
	return query
}
