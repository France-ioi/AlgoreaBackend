// +build !prod

package testhelpers

import (
	"github.com/cucumber/messages-go/v10"
)

func (ctx *TestContext) TheTemplateConstantIsString(name, value string) error { // nolint
	ctx.templateSet.AddGlobal(name, value)
	return nil
}

func (ctx *TestContext) TheTemplateConstantIsDocString(name string, value *messages.PickleStepArgument_PickleDocString) error { // nolint
	ctx.templateSet.AddGlobal(name, value.Content)
	return nil
}
