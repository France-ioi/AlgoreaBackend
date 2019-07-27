package testhelpers

import (
	"github.com/DATA-DOG/godog/gherkin"
)

func (ctx *TestContext) TheTemplateConstantIsString(name, value string) error { // nolint
	ctx.templateSet.AddGlobal(name, value)
	return nil
}

func (ctx *TestContext) TheTemplateConstantIsDocString(name string, value *gherkin.DocString) error { // nolint
	ctx.templateSet.AddGlobal(name, value.Content)
	return nil
}
