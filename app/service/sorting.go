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
	// Nullable means that the field can be null
	Nullable bool
	// Unique means that sorting rules containing this parameter will not be augmented with a tie-breaker field
	Unique bool
}

type sortingDirection int

const (
	asc  sortingDirection = 1
	desc sortingDirection = -1
)

type nullPlacement int

const (
	first nullPlacement = -1
	last  nullPlacement = 1
)

type sortingType struct {
	sortingDirection
	nullPlacement
}

func (t sortingType) asSQL(columnName string) string {
	var result string
	switch t.nullPlacement {
	case last:
		result += columnName + " IS NULL, "
	case first:
		result += columnName + " IS NOT NULL, "
	}
	result += columnName + " "
	if t.sortingDirection == asc {
		return result + "ASC"
	}
	return result + "DESC"
}

func (t sortingType) conditionSign() string {
	if t.sortingDirection == asc {
		return ">"
	}
	return "<"
}

// ApplySortingAndPaging applies ordering and paging according to given accepted fields and sorting types
// taking into the account the URL parameters 'from.*'.
// When the `skipSortParameter` is true, the 'sort' request parameter is ignored.
func ApplySortingAndPaging(r *http.Request, query *database.DB, acceptedFields map[string]*FieldSortingParams,
	defaultRules string, tieBreakerFieldNames []string, skipSortParameter bool) (*database.DB, APIError) {
	sortingRules := prepareSortingRulesAndAcceptedFields(r, defaultRules, skipSortParameter)

	usedFields, fieldsSortingTypes, err := parseSortingRules(sortingRules, acceptedFields, tieBreakerFieldNames)
	if err != nil {
		return nil, ErrInvalidRequest(err)
	}
	query = applyOrder(query, usedFields, acceptedFields, fieldsSortingTypes)

	fromValues, err := parsePagingParameters(r, usedFields, acceptedFields, fieldsSortingTypes)
	if err != nil {
		return nil, ErrInvalidRequest(err)
	}

	query = applyPagingConditions(query, usedFields, fieldsSortingTypes, acceptedFields, fromValues)
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

// parseSortingRules returns a slice with used fields and a map fieldName -> sortingType
// It also checks that there are no unallowed fields in the rules.
func parseSortingRules(sortingRules string, acceptedFields map[string]*FieldSortingParams, tieBreakerFieldNames []string) (
	usedFields []string, fieldsSortingTypes map[string]sortingType, err error) {
	sortStatements := strings.Split(sortingRules, ",")
	usedFields = make([]string, 0, len(sortStatements)+1)
	fieldsSortingTypes = make(map[string]sortingType, len(sortStatements)+1)
	includesUniqueField := false
	for _, sortStatement := range sortStatements {
		fieldName, sorting := getFieldNameAndSortingTypeFromSortStatement(sortStatement)
		err = validateSortingField(fieldsSortingTypes, fieldName, acceptedFields, sorting)
		if err != nil {
			return nil, nil, err
		}
		if acceptedFields[fieldName].Nullable && sorting.nullPlacement == 0 {
			sorting.nullPlacement = first
		}
		fieldsSortingTypes[fieldName] = sorting
		usedFields = append(usedFields, fieldName)
		if acceptedFields[fieldName].Unique {
			includesUniqueField = true
		}
	}
	if !includesUniqueField {
		for _, tieBreakerFieldName := range tieBreakerFieldNames {
			if fieldsSortingTypes[tieBreakerFieldName].sortingDirection == 0 {
				fieldsSortingTypes[tieBreakerFieldName] = sortingType{sortingDirection: 1}
				usedFields = append(usedFields, tieBreakerFieldName)
			}
		}
	}
	return usedFields, fieldsSortingTypes, err
}

func validateSortingField(
	fieldsSortingTypes map[string]sortingType, fieldName string, acceptedFields map[string]*FieldSortingParams, sorting sortingType) error {
	if _, ok := fieldsSortingTypes[fieldName]; ok {
		return fmt.Errorf("a field cannot be a sorting parameter more than once: %q", fieldName)
	}
	if _, ok := acceptedFields[fieldName]; !ok {
		return fmt.Errorf("unallowed field in sorting parameters: %q", fieldName)
	}
	if !acceptedFields[fieldName].Nullable && sorting.nullPlacement != 0 {
		return fmt.Errorf("'null last' sorting cannot be applied to a non-nullable field: %q", fieldName)
	}
	return nil
}

// getFieldNameAndSortingTypeFromSortStatement extracts a field name and a sorting type
// from a given sorting statement.
// "id"   -> ("id", 1)
// "-name -> ("name", {-1, 0})
// "-date$ -> ("date", {-1, 1}) # null last
func getFieldNameAndSortingTypeFromSortStatement(sortStatement string) (string, sortingType) {
	var direction sortingDirection
	var np nullPlacement
	if len(sortStatement) > 0 && sortStatement[0] == '-' {
		sortStatement = sortStatement[1:]
		direction = desc
	} else {
		direction = asc
	}
	if len(sortStatement) > 0 && sortStatement[len(sortStatement)-1] == '$' {
		np = last
		sortStatement = sortStatement[:len(sortStatement)-1]
	}
	fieldName := sortStatement
	return fieldName, sortingType{direction, np}
}

// applyOrder appends the "ORDER BY" statement to given query according to the given list of used fields,
// the fields configuration (acceptedFields) and sorting types
func applyOrder(query *database.DB, usedFields []string, acceptedFields map[string]*FieldSortingParams,
	fieldsSortingTypes map[string]sortingType) *database.DB {
	usedFieldsNumber := len(usedFields)
	orderStrings := make([]string, 0, usedFieldsNumber)
	for _, field := range usedFields {
		var columnName string
		if acceptedFields[field].ColumnNameForOrdering != "" {
			columnName = acceptedFields[field].ColumnNameForOrdering
		} else {
			columnName = acceptedFields[field].ColumnName
		}
		orderStrings = append(orderStrings, fieldsSortingTypes[field].asSQL(columnName))
	}
	if len(orderStrings) > 0 {
		query = query.Order(strings.Join(orderStrings, ", "))
	}
	return query
}

// parsePagingParameters returns a slice of values provided for paging in a request URL (none or all should be present)
// The values are in the order of the 'usedFields' and converted according to 'FieldType' of 'acceptedFields'
func parsePagingParameters(r *http.Request, usedFields []string,
	acceptedFields map[string]*FieldSortingParams, fieldsSortingTypes map[string]sortingType) ([]interface{}, error) {
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
			if _, ok := fieldsSortingTypes[fieldNameSuffix]; !ok {
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
	if acceptedFields[fieldName].Nullable && r.URL.Query()[fromFieldName][0] == "[NULL]" {
		return nil, nil
	}
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
func applyPagingConditions(query *database.DB, usedFields []string, fieldsSortingTypes map[string]sortingType,
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
		columnName := acceptedFields[fieldName].ColumnName
		if fromValues[index] == nil {
			if fieldsSortingTypes[fieldName].nullPlacement == first {
				conditions = append(conditions,
					fmt.Sprintf("(%s%s IS NOT NULL)", conditionPrefix, columnName))
			} else if index == len(usedFields)-1 {
				conditions = append(conditions,
					fmt.Sprintf("(%s%s IS NULL)", conditionPrefix, columnName))
			}
			conditionPrefix = fmt.Sprintf("%s%s IS NULL", conditionPrefix, columnName)
		} else {
			condition := fmt.Sprintf("%s %s ?", columnName, fieldsSortingTypes[fieldName].conditionSign())
			if fieldsSortingTypes[fieldName].nullPlacement == last {
				condition = fmt.Sprintf("(%s OR %s IS NULL)", condition, columnName)
			}
			conditions = append(conditions, fmt.Sprintf("(%s%s)", conditionPrefix, condition))
			conditionPrefix = fmt.Sprintf("%s%s = ?", conditionPrefix, columnName)

			queryValuesPart = append(queryValuesPart, fromValues[index])
			queryValues = append(queryValues, queryValuesPart...)
		}
	}
	if len(conditions) > 0 {
		query = query.Where(strings.Join(conditions, " OR "), queryValues...)
	}
	return query
}
