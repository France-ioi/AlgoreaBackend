package service

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/France-ioi/AlgoreaBackend/app/database"
)

// FieldSortingParams represents sorting parameters for one field
type FieldSortingParams struct {
	// ColumnName is a DB column name (may contain a table name as a prefix, e.g. "groups.id")
	ColumnName string
	// ColumnNameForOrdering is a DB column name (may contain a table name as a prefix, e.g. "groups.id")
	// used (if set) in ORDER BY clause instead of 'ColumnName'
	ColumnNameForOrdering string
	// FieldType is one of "int64", "bool", "string", "time"
	FieldType string
	// Unique means that sorting rules containing this parameter will not be augmented with a tie-breaker field
	Unique bool
}

type sortingDirection int

const (
	asc     sortingDirection = 1
	desc    sortingDirection = -1
	unknown sortingDirection = 0
)

func (d sortingDirection) asSQL() string {
	if d == asc {
		return "ASC"
	}
	return "DESC"
}

func (d sortingDirection) conditionSign() string {
	if d == asc {
		return ">"
	}
	return "<"
}

// ApplySortingAndPaging applies ordering and paging according to given accepted fields and sorting rules
// taking into the account the URL parameters 'from.*'.
// When the `skipSortParameter` is true, the 'sort' request parameter is ignored.
func ApplySortingAndPaging(r *http.Request, query *database.DB, acceptedFields map[string]*FieldSortingParams,
	defaultRules, tieBreakerFieldName string, skipSortParameter bool) (*database.DB, APIError) {
	sortingRules := prepareSortingRulesAndAcceptedFields(r, defaultRules, skipSortParameter)

	usedFields, fieldsDirections, err := parseSortingRules(sortingRules, acceptedFields, tieBreakerFieldName)
	if err != nil {
		return nil, ErrInvalidRequest(err)
	}
	query = applyOrder(query, usedFields, acceptedFields, fieldsDirections)

	fromValues, err := parsePagingParameters(r, usedFields, acceptedFields, fieldsDirections)
	if err != nil {
		return nil, ErrInvalidRequest(err)
	}

	query = applyPagingConditions(query, usedFields, fieldsDirections, acceptedFields, fromValues)
	return query, NoError
}

// prepareSortingRulesAndAcceptedFields builds sorting rules and a map of accepted fields.
// If urlQuery["sort"] is not present, the default sorting rules are used.
func prepareSortingRulesAndAcceptedFields(r *http.Request, defaultRules string, skipSortParameter bool) (sortingRules string) {
	urlQuery := r.URL.Query()
	if !skipSortParameter && len(urlQuery["sort"]) > 0 {
		sortingRules = urlQuery["sort"][0]
	} else {
		sortingRules = defaultRules
	}
	return sortingRules
}

// parseSortingRules returns a slice with used fields and a map fieldName -> direction
// It also checks that there are no unallowed fields in the rules.
func parseSortingRules(sortingRules string, acceptedFields map[string]*FieldSortingParams, tieBreakerFieldName string) (
	usedFields []string, fieldsDirections map[string]sortingDirection, err error) {
	sortStatements := strings.Split(sortingRules, ",")
	usedFields = make([]string, 0, len(sortStatements)+1)
	fieldsDirections = make(map[string]sortingDirection, len(sortStatements)+1)
	includesUniqueField := false
	for _, sortStatement := range sortStatements {
		fieldName, direction := getFieldNameAndDirectionFromSortStatement(sortStatement)
		if fieldsDirections[fieldName] != 0 {
			return nil, nil, fmt.Errorf("a field cannot be a sorting parameter more than once: %q", fieldName)
		}
		if _, ok := acceptedFields[fieldName]; !ok {
			return nil, nil, fmt.Errorf("unallowed field in sorting parameters: %q", fieldName)
		}
		fieldsDirections[fieldName] = direction
		usedFields = append(usedFields, fieldName)
		if acceptedFields[fieldName].Unique {
			includesUniqueField = true
		}
	}
	if !includesUniqueField {
		if fieldsDirections[tieBreakerFieldName] == 0 {
			fieldsDirections[tieBreakerFieldName] = 1
			usedFields = append(usedFields, tieBreakerFieldName)
		}
	}
	return
}

// getFieldNameAndDirectionFromSortStatement extracts a field name and a sorting direction
// from a given sorting statement.
// "id"   -> ("id", 1)
// "-name -> ("name", -1)
func getFieldNameAndDirectionFromSortStatement(sortStatement string) (string, sortingDirection) {
	var direction sortingDirection
	if len(sortStatement) > 0 && sortStatement[0] == '-' {
		sortStatement = sortStatement[1:]
		direction = desc
	} else {
		direction = asc
	}
	fieldName := sortStatement
	return fieldName, direction
}

// applyOrder appends the "ORDER BY" statement to given query according to the given list of used fields,
// the fields configuration (acceptedFields) and sorting directions
func applyOrder(query *database.DB, usedFields []string, acceptedFields map[string]*FieldSortingParams,
	fieldsDirections map[string]sortingDirection) *database.DB {
	usedFieldsNumber := len(usedFields)
	orderStrings := make([]string, 0, usedFieldsNumber)
	for _, field := range usedFields {
		var columnName string
		if acceptedFields[field].ColumnNameForOrdering != "" {
			columnName = acceptedFields[field].ColumnNameForOrdering
		} else {
			columnName = acceptedFields[field].ColumnName
		}
		orderStrings = append(orderStrings, columnName+" "+fieldsDirections[field].asSQL())
	}
	if len(orderStrings) > 0 {
		query = query.Order(strings.Join(orderStrings, ", "))
	}
	return query
}

// parsePagingParameters returns a slice of values provided for paging in a request URL (none or all should be present)
// The values are in the order of the 'usedFields' and converted according to 'FieldType' of 'acceptedFields'
func parsePagingParameters(r *http.Request, usedFields []string,
	acceptedFields map[string]*FieldSortingParams, fieldsDirections map[string]sortingDirection) ([]interface{}, error) {
	fromValueSkipped := false
	fromValueAccepted := false
	fromValues := make([]interface{}, 0, len(usedFields))
	const fromPrefix = "from."
	for _, fieldName := range usedFields {
		var value interface{}
		fromFieldName := fromPrefix + fieldName
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

	var unknownFromFields []string
	for fieldName := range r.URL.Query() {
		if strings.HasPrefix(fieldName, fromPrefix) {
			fieldNameSuffix := fieldName[len(fromPrefix):]
			if fieldsDirections[fieldNameSuffix] == unknown {
				unknownFromFields = append(unknownFromFields, fieldName)
			}
		}
	}
	if len(unknownFromFields) > 0 {
		sort.Strings(unknownFromFields)
		return nil, fmt.Errorf("unallowed paging parameters (%s)", strings.Join(unknownFromFields, ", "))
	}

	return fromValues, nil
}

// getFromValueForField returns a 'from' value (a paging parameter) for a given field.
// The value is converted according to 'FieldType' of 'acceptedFields[fieldName]'
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
	case "time":
		return ResolveURLQueryGetTimeField(r, fromFieldName)
	default:
		panic(fmt.Errorf("unsupported type %q for field %q", acceptedFields[fieldName].FieldType, fieldName))
	}
}

// applyPagingConditions adds filtering on paging values into the query
func applyPagingConditions(query *database.DB, usedFields []string, fieldsDirections map[string]sortingDirection,
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
		conditions = append(conditions,
			fmt.Sprintf("(%s%s %s ?)", conditionPrefix, acceptedFields[fieldName].ColumnName,
				fieldsDirections[fieldName].conditionSign()))
		conditionPrefix = fmt.Sprintf("%s%s = ?", conditionPrefix, acceptedFields[fieldName].ColumnName)

		queryValuesPart = append(queryValuesPart, fromValues[index])
		queryValues = append(queryValues, queryValuesPart...)
	}
	if len(conditions) > 0 {
		query = query.Where(strings.Join(conditions, " OR "), queryValues...)
	}
	return query
}
