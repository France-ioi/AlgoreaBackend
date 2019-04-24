package types

type SampleInt32Input struct {
	Required         RequiredInt32
	Nullable         NullableInt32
	Optional         OptionalInt32
	OptionalNullable OptNullInt32
}

func (v *SampleInt32Input) Validate() error {
	return Validate([]string{"Required", "Nullable", "Optional", "OptionalNullable"},
		&v.Required, &v.Nullable, &v.Optional, &v.OptionalNullable)
}
