package service

import (
	"strconv"
	"strings"
	"time"
)

// ConvertSliceOfMapsFromDBToJSON given a slice of maps that represents DB result data,
// converts it to a slice of maps for rendering JSON so that:
// 1) all maps keys with "__" are considered as paths in JSON (converts "User__ID":... to "user":{"id": ...})
// 2) all maps keys are converted to snake case
// 3) prefixes are stripped, values are converted to needed types accordingly.
func ConvertSliceOfMapsFromDBToJSON(dbMaps []map[string]any) []map[string]any {
	convertedResult := make([]map[string]any, len(dbMaps))
	for index := range dbMaps {
		convertedResult[index] = ConvertMapFromDBToJSON(dbMaps[index])
	}
	return convertedResult
}

// ConvertMapFromDBToJSON given a map that represents DB result data,
// converts it a map for rendering JSON so that:
// 1) all map keys with "__" are considered as paths in JSON (converts "User__ID":... to "user":{"id": ...})
// 2) all map keys are converted to snake case
// 3) prefixes are stripped, values are converted to needed types accordingly.
func ConvertMapFromDBToJSON(dbMap map[string]any) map[string]any {
	result := map[string]any{}
	for key, value := range dbMap {
		currentMap := result

		subKeys := strings.Split(key, "__")
		for subKeyIndex, subKey := range subKeys {
			if subKeyIndex == len(subKeys)-1 {
				setConvertedValueToJSONMap(subKey, value, currentMap)
				continue
			}
			shouldCreateSubMap := true
			if subMap, hasSubMap := currentMap[subKey]; hasSubMap {
				if subMap, ok := subMap.(map[string]any); ok {
					currentMap = subMap
					shouldCreateSubMap = false
				}
			}
			if shouldCreateSubMap {
				newMap := map[string]any{}
				currentMap[subKey] = newMap
				currentMap = newMap
			}
		}
	}

	replaceEmptySubMapsWithNils(result)
	return result
}

func replaceEmptySubMapsWithNils(mapToProcess map[string]any) bool {
	for key := range mapToProcess {
		if subMap, ok := mapToProcess[key].(map[string]any); ok {
			if replaceEmptySubMapsWithNils(subMap) {
				mapToProcess[key] = nil
			}
		}
	}
	for key := range mapToProcess {
		if mapToProcess[key] != nil {
			return false
		}
	}
	return true
}

func setConvertedValueToJSONMap(valueName string, value any, result map[string]any) {
	if valueInt64, ok := value.(int64); ok {
		value = strconv.FormatInt(valueInt64, 10)
	}

	value = convertTimeToRFC3339IfTime(value, valueName)
	result[valueName] = value
}

func convertTimeToRFC3339IfTime(value any, snakeCaseName string) any {
	if value != nil &&
		(strings.HasSuffix(snakeCaseName, "_date") || strings.HasSuffix(snakeCaseName, "_at") ||
			strings.HasSuffix(snakeCaseName, "_since") || strings.HasSuffix(snakeCaseName, "_until") ||
			snakeCaseName == "at") {
		value = ConvertDBTimeToJSONTime(value)
	}
	return value
}

// ConvertDBTimeToJSONTime converts the DB datetime representation to RFC3339.
func ConvertDBTimeToJSONTime(data any) string {
	//nolint:forcetypeassert // panic if the data is not a string
	parsedTime, err := time.Parse("2006-01-02 15:04:05.999", data.(string))
	if err != nil {
		panic(err)
	}
	return parsedTime.Format(time.RFC3339Nano)
}
