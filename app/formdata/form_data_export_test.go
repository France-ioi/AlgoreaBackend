package formdata

import "github.com/France-ioi/mapstructure"

func ToAnythingHookFunc() mapstructure.DecodeHookFunc {
	return toAnythingHookFunc()
}

func StringToInt64HookFunc() mapstructure.DecodeHookFunc {
	return stringToInt64HookFunc()
}
