package types

const emptyJSONStruct = `{}`
const allNullsJSONStruct = `{ "Required": null, "Nullable": null, "Optional": null, "OptionalNullable": null }`

type SampleBoolInput struct {
	Required         RequiredBool
	Nullable         NullableBool
	Optional         OptionalBool
	OptionalNullable OptNullBool
}

func (v *SampleBoolInput) Validate() error {
	return Validate([]string{"Required", "Nullable", "Optional", "OptionalNullable"},
		&v.Required, &v.Nullable, &v.Optional, &v.OptionalNullable)
}
