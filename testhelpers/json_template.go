package testhelpers

import (
	"bytes"
	"crypto/rsa"
	"fmt"
	"io"
	"reflect"
	"regexp"
	"strconv"
	"time"

	"github.com/CloudyKit/jet"

	"github.com/France-ioi/AlgoreaBackend/app/token"
)

var dbPathRegexp = regexp.MustCompile(`^\s*(\w+)\[(\d+)]\[(\w+)]\s*$`)

func (ctx *TestContext) preprocessJSONBody(jsonBody string) (string, error) {
	set := jet.NewSet(jet.SafeWriter(func(w io.Writer, b []byte) {
		w.Write(b) // nolint:gosec,errcheck
	}))
	set.AddGlobalFunc("currentTimeInFormat", func(a jet.Arguments) reflect.Value {
		a.RequireNumOfArguments("currentTimeInFormat", 1, 1)
		return reflect.ValueOf(time.Now().UTC().Format(a.Get(0).Interface().(string)))
	})
	set.AddGlobalFunc("generateToken", func(a jet.Arguments) reflect.Value {
		a.RequireNumOfArguments("generateToken", 2, 2)
		return reflect.ValueOf(
			fmt.Sprintf("%q", token.Generate(a.Get(0).Interface().(map[string]interface{}),
				a.Get(1).Addr().Interface().(*rsa.PrivateKey))))
	})
	set.AddGlobalFunc("app", func(a jet.Arguments) reflect.Value {
		return reflect.ValueOf(ctx.application)
	})
	set.AddGlobalFunc("db", func(a jet.Arguments) reflect.Value {
		a.RequireNumOfArguments("db", 1, 1)
		path := a.Get(0).Interface().(string)
		if match := dbPathRegexp.FindStringSubmatch(path); match != nil {
			gherkinTable := ctx.dbTableData[match[1]]
			neededColumnNumber := -1
			for columnNumber, cell := range gherkinTable.Rows[0].Cells {
				if cell.Value == match[3] {
					neededColumnNumber = columnNumber
					break
				}
			}
			if neededColumnNumber == -1 {
				a.Panicf("cannot find column %q in table %q", match[3], match[1])
			}
			rowNumber, conversionErr := strconv.Atoi(match[2])
			if conversionErr != nil {
				a.Panicf("can't convert a row number: %s", conversionErr.Error())
			}
			return reflect.ValueOf(gherkinTable.Rows[rowNumber].Cells[neededColumnNumber].Value)
		}
		a.Panicf("wrong data path: %q", path)
		return reflect.Value{}
	})
	tmpl, err := set.LoadTemplate("template", jsonBody)

	if err != nil {
		return "", err
	}
	buffer := bytes.NewBuffer(make([]byte, 0, 1024))
	err = tmpl.Execute(buffer, nil, nil)
	if err != nil {
		return "", err
	}
	jsonBody = buffer.String()

	return jsonBody, nil
}
