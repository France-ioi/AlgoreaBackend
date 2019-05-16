package formdata

import "github.com/France-ioi/mapstructure"

func ToAnythingHookFunc() mapstructure.DecodeHookFunc {
	return toAnythingHookFunc()
}
