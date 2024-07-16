//go:build !prod

package testhelpers

import (
	"bytes"
	"crypto/rsa"
	"fmt"
	"io"
	"math/rand"
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
	dbPathRegexp            = regexp.MustCompile(`^\s*(\w+)\[(\d+)]\[(\w+)]\s*$`)
	replaceReferencesRegexp = regexp.MustCompile(`(^|\W)(@\w+)`)
)

// getOrCreateReferenceFor gets the ID from a reference, or create the reference if it doesn't exist.
func (ctx *TestContext) getReference(reference string) int64 {
	if id, err := strconv.ParseInt(reference, 10, 64); err == nil {
		return id
	}

	if value, ok := ctx.identifierReferences[reference]; ok {
		return value
	}

	id := rand.Int63()
	ctx.identifierReferences[reference] = id

	return id
}

// replaceReferencesByIDs changes the references (@ref) in a string by the referenced identifiers (ID).
func (ctx *TestContext) replaceReferencesByIDs(str string) string {
	// a reference should either be at the beginning of the string (^), or after a non alpha-num character (\W).
	// we don't want to rewrite email addresses.
	return replaceReferencesRegexp.ReplaceAllStringFunc(str, func(capture string) string {
		// capture is either:
		// - @Reference
		// - /@Reference (or another non-alphanum character in front)

		if capture[0] == ReferencePrefix {
			return strconv.FormatInt(ctx.getReference(capture), 10)
		}

		return string(capture[0]) + strconv.FormatInt(ctx.getReference(capture[1:]), 10)
	})
}

func (ctx *TestContext) preprocessString(jsonBody string) (string, error) {
	jsonBody = ctx.replaceReferencesByIDs(jsonBody)
	tmpl, err := ctx.templateSet.Parse("template", jsonBody)
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

func (ctx *TestContext) constructTemplateSet() *jet.Set {
	set := jet.NewSet(jet.SafeWriter(func(w io.Writer, b []byte) {
		w.Write(b) //nolint:gosec,errcheck
	}))

	set.AddGlobalFunc("currentTimeInFormat", func(a jet.Arguments) reflect.Value {
		a.RequireNumOfArguments("currentTimeInFormat", 1, 1)
		return reflect.ValueOf(time.Now().UTC().Format(a.Get(0).Interface().(string)))
	})

	set.AddGlobalFunc("generateToken", func(a jet.Arguments) reflect.Value {
		a.RequireNumOfArguments("generateToken", 2, 2)
		var privateKey *rsa.PrivateKey
		privateKeyRefl := a.Get(1)
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
				a.Panicf("Cannot parse private key: %s", err)
			}
		}
		return reflect.ValueOf(
			fmt.Sprintf("%q", token.Generate(a.Get(0).Interface().(map[string]interface{}),
				privateKey)))
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

	set.AddGlobalFunc("relativeTime", func(a jet.Arguments) reflect.Value {
		a.RequireNumOfArguments("relativeTime", 1, 1)
		durationString := a.Get(0).Interface().(string)
		duration, err := time.ParseDuration(durationString)
		if err != nil {
			a.Panicf("can't parse duration: %s", err.Error())
		}
		return reflect.ValueOf(time.Now().UTC().Add(duration).Format("2006-01-02 15:04:05"))
	})

	addTimeToRFCFunction(set)

	set.AddGlobal("taskPlatformPublicKey", tokentest.TaskPlatformPublicKey)
	set.AddGlobal("taskPlatformPrivateKey", tokentest.TaskPlatformPrivateKey)

	return set
}

func addTimeToRFCFunction(set *jet.Set) {
	set.AddGlobalFunc("timeToRFC", func(a jet.Arguments) reflect.Value {
		a.RequireNumOfArguments("timeToRFC", 1, 1)
		dbTime := a.Get(0).Interface().(string)
		parsedTime, err := time.Parse("2006-01-02 15:04:05", dbTime)
		if err != nil {
			a.Panicf("can't parse mysql datetime: %s", err.Error())
		}
		return reflect.ValueOf(parsedTime.Format(time.RFC3339))
	})
}
