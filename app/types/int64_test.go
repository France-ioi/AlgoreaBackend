package types

type SampleInt64Input struct {
	Required         RequiredInt64
	Nullable         NullableInt64
	Optional         OptionalInt64
	OptionalNullable OptNullInt64
}

func (v *SampleInt64Input) Validate() error {
	return Validate([]string{"Required", "Nullable", "Optional", "OptionalNullable"},
		&v.Required, &v.Nullable, &v.Optional, &v.OptionalNullable)
}
