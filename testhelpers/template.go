//go:build !prod

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
	"github.com/SermoDigital/jose/crypto"

	"github.com/France-ioi/AlgoreaBackend/v2/app/token"
	"github.com/France-ioi/AlgoreaBackend/v2/app/tokentest"
)

var (
	dbPathRegexp    = regexp.MustCompile(`^\s*(\w+)\[(\d+)]\[(\w+)]\s*$`)
	referenceRegexp = regexp.MustCompile(`(^|\W)(@\w+)`)
)

// getIDOfReference returns the ID of a reference.
func (ctx *TestContext) getIDOfReference(reference string) int64 {
	if id, err := strconv.ParseInt(reference, 10, 64); err == nil {
		return id
	}

	if value, ok := ctx.referenceToIDMap[reference]; ok {
		return value
	}

	panic(fmt.Sprintf("reference %q not found", reference))
}

// replaceReferencesWithIDs changes the references (@ref) in a string with the referenced identifiers (ID).
func (ctx *TestContext) replaceReferencesWithIDs(str string) string {
	// a reference should either be at the beginning of the string (^), or after a non alpha-num character (\W).
	// we don't want to rewrite email addresses.
	return referenceRegexp.ReplaceAllStringFunc(str, func(capture string) string {
		// capture is either:
		// - @Reference
		// - /@Reference (or another non-alphanum character in front)

		if capture[0] == referencePrefix {
			return strconv.FormatInt(ctx.getIDOfReference(capture), 10)
		}

		return string(capture[0]) + strconv.FormatInt(ctx.getIDOfReference(capture[1:]), 10)
	})
}

func (ctx *TestContext) preprocessString(str string) string {
	str = ctx.replaceReferencesWithIDs(str)
	tmpl, err := ctx.templateSet.Parse("template", str)
	mustNotBeError(err)
	const kilobyte = 1024
	buffer := bytes.NewBuffer(make([]byte, 0, kilobyte))
	err = tmpl.Execute(buffer, nil, nil)
	mustNotBeError(err)
	return buffer.String()
}

func (ctx *TestContext) constructTemplateSet() *jet.Set {
	set := jet.NewSet(jet.SafeWriter(func(w io.Writer, b []byte) {
		w.Write(b) //nolint:gosec,errcheck
	}))

	set.AddGlobalFunc("currentTimeInFormat", func(a jet.Arguments) reflect.Value {
		a.RequireNumOfArguments("currentTimeInFormat", 1, 1)
		return reflect.ValueOf(time.Now().UTC().Format(a.Get(0).Interface().(string)))
	})

	set.AddGlobalFunc("currentTimeDB", func(_ jet.Arguments) reflect.Value {
		return reflect.ValueOf(time.Now().UTC().Truncate(time.Microsecond).Format("2006-01-02 15:04:05.999999"))
	})

	set.AddGlobalFunc("currentTimeDBMs", func(_ jet.Arguments) reflect.Value {
		return reflect.ValueOf(time.Now().UTC().Truncate(time.Millisecond).Format("2006-01-02 15:04:05.000"))
	})

	set.AddGlobalFunc("generateToken", func(arguments jet.Arguments) reflect.Value {
		//nolint:mnd // we require exactly two arguments: the payload and the private key
		arguments.RequireNumOfArguments("generateToken", 2, 2)
		var privateKey *rsa.PrivateKey
		privateKeyRefl := arguments.Get(1)
		if privateKeyRefl.CanAddr() {
			privateKey = privateKeyRefl.Addr().Interface().(*rsa.PrivateKey)
		} else {
			var err error
			var privateKeyBytes []byte
			if privateKeyRefl.Kind() == reflect.String {
				privateKeyBytes = []byte(privateKeyRefl.Interface().(string))
			} else {
				privateKeyBytes = privateKeyRefl.Interface().([]byte)
			}
			privateKey, err = crypto.ParseRSAPrivateKeyFromPEM(privateKeyBytes)
			if err != nil {
				arguments.Panicf("Cannot parse private key: %s", err)
			}
		}
		return reflect.ValueOf(
			fmt.Sprintf("%q", token.Generate(arguments.Get(0).Interface().(map[string]interface{}),
				privateKey)))
	})

	set.AddGlobalFunc("app", func(_ jet.Arguments) reflect.Value {
		return reflect.ValueOf(ctx.application)
	})

	set.AddGlobalFunc("db", func(arguments jet.Arguments) reflect.Value {
		arguments.RequireNumOfArguments("db", 1, 1)
		path := arguments.Get(0).Interface().(string)
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
				arguments.Panicf("cannot find column %q in table %q", match[3], match[1])
			}
			rowNumber, conversionErr := strconv.Atoi(match[2])
			if conversionErr != nil {
				arguments.Panicf("can't convert a row number: %s", conversionErr.Error())
			}
			return reflect.ValueOf(gherkinTable.Rows[rowNumber].Cells[neededColumnNumber].Value)
		}
		arguments.Panicf("wrong data path: %q", path)
		return reflect.Value{}
	})

	set.AddGlobalFunc("relativeTimeDB", func(a jet.Arguments) reflect.Value {
		a.RequireNumOfArguments("relativeTimeDB", 1, 1)
		durationString := a.Get(0).Interface().(string)
		duration, err := time.ParseDuration(durationString)
		if err != nil {
			a.Panicf("can't parse duration: %s", err.Error())
		}
		return reflect.ValueOf(time.Now().UTC().Add(duration).Truncate(time.Second).Format(time.DateTime))
	})

	addRelativeTimeDBMsFunction(set)
	addtimeDBToRFC3339Function(set)
	addTimeDBMsToRFC3339Function(set)

	set.AddGlobal("taskPlatformPublicKey", tokentest.TaskPlatformPublicKey)
	set.AddGlobal("taskPlatformPrivateKey", tokentest.TaskPlatformPrivateKey)

	set.AddGlobalFunc("quote", func(a jet.Arguments) reflect.Value {
		a.RequireNumOfArguments("quote", 1, 1)
		return reflect.ValueOf(fmt.Sprintf("%q", a.Get(0).Interface()))
	})

	return set
}

func addRelativeTimeDBMsFunction(set *jet.Set) *jet.Set {
	return set.AddGlobalFunc("relativeTimeDBMs", func(a jet.Arguments) reflect.Value {
		a.RequireNumOfArguments("relativeTimeDBMs", 1, 1)
		durationString := a.Get(0).Interface().(string)
		duration, err := time.ParseDuration(durationString)
		if err != nil {
			a.Panicf("can't parse duration: %s", err.Error())
		}
		return reflect.ValueOf(time.Now().UTC().Add(duration).Truncate(time.Millisecond).Format("2006-01-02 15:04:05.000"))
	})
}

func addtimeDBToRFC3339Function(set *jet.Set) {
	set.AddGlobalFunc("timeDBToRFC3339", func(a jet.Arguments) reflect.Value {
		a.RequireNumOfArguments("timeDBToRFC3339", 1, 1)
		dbTime := a.Get(0).Interface().(string)
		parsedTime, err := time.Parse("2006-01-02 15:04:05.999999999", dbTime)
		if err != nil {
			a.Panicf("can't parse mysql datetime: %s", err.Error())
		}
		parsedTime = parsedTime.Truncate(time.Second)
		return reflect.ValueOf(parsedTime.Format(time.RFC3339))
	})
}

func addTimeDBMsToRFC3339Function(set *jet.Set) {
	set.AddGlobalFunc("timeDBMsToRFC3339", func(a jet.Arguments) reflect.Value {
		a.RequireNumOfArguments("timeDBMsToRFC3339", 1, 1)
		dbTime := a.Get(0).Interface().(string)
		parsedTime, err := time.Parse("2006-01-02 15:04:05.999", dbTime)
		if err != nil {
			a.Panicf("can't parse mysql datetime: %s", err.Error())
		}
		return reflect.ValueOf(parsedTime.Format(time.RFC3339Nano))
	})
}
