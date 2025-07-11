//go:build !prod

package testhelpers

import (
	"github.com/cucumber/godog"
)

func (ctx *TestContext) TheTemplateConstantIsString(name, value string) error { //nolint
	value = ctx.preprocessString(value)
	ctx.templateSet.AddGlobal(name, value)
	return nil
}

func (ctx *TestContext) TheTemplateConstantIsDocString(name string, value *godog.DocString) error { //nolint
	preprocessedValue := ctx.preprocessString(value.Content)
	ctx.templateSet.AddGlobal(name, preprocessedValue)
	return nil
}
