// +build !prod

package testhelpers

import (
	"github.com/cucumber/messages-go/v10"
)

func (ctx *TestContext) TheTemplateConstantIsString(name, value string) error { // nolint
	value, err := ctx.preprocessString(value)
	if err != nil {
		return err
	}

	ctx.templateSet.AddGlobal(name, value)
	return nil
}

func (ctx *TestContext) TheTemplateConstantIsDocString(name string, value *messages.PickleStepArgument_PickleDocString) error { // nolint
	preprocessedValue, err := ctx.preprocessString(value.Content)
	if err != nil {
		return err
	}

	ctx.templateSet.AddGlobal(name, preprocessedValue)
	return nil
}
