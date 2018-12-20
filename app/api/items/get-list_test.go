package items

import (
	"encoding/json"
	"github.com/France-ioi/AlgoreaBackend/app/config"
	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
	"github.com/jinzhu/gorm"
	assert_lib "github.com/stretchr/testify/assert"
	"log"
	"os"
	"os/exec"
	"testing"
)

func TestIsValidHierarchy(t *testing.T) {
	assert := assert_lib.New(t)

	config.Path = "testdata/default.yaml"
	conf, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}

	db, err := database.Open(conf.Database)
	if err != nil {
		t.Fatal(err)
	}

	db.SetLogger(gorm.Logger{log.New(os.Stdout, "\r\n", 0)})
	db.LogMode(true)

	base := service.Base{Store: database.NewDataStore(db), Config: conf}
	srv := &Service{Base: base}

	testCases := []struct {
		desc   string
		items  []string
		args   []int64
		expect bool
	}{
		{
			desc: "normal",
			items: []string{
				`{"id":21, "type":"Root", "strings":[{"language_id":1, "title":"21"}], "parents": [{"id":21, "order":0}]}`,
				`{"id":22, "type":"Category", "strings":[{"language_id":1, "title":"22"}], "parents": [{"id":21, "order":0}]}`,
				`{"id":23, "type":"Chapter", "strings":[{"language_id":1, "title":"23"}], "parents": [{"id":22, "order":1}]}`,
			},
			args:   []int64{21, 22, 23},
			expect: true,
		},
		{
			desc: "one item missing",
			items: []string{
				`{"id":21, "type":"Root", "strings":[{"language_id":1, "title":"21"}], "parents": [{"id":21, "order":0}]}`,
				`{"id":22, "type":"Category", "strings":[{"language_id":1, "title":"22"}], "parents": [{"id":21, "order":0}]}`,
				// `{"id":23, "type":"Chapter", "strings":[{"language_id":1, "title":"23"}], "parents": [{"id":22, "order":1}]}`,
				`{"id":24, "type":"Chapter", "strings":[{"language_id":1, "title":"23"}], "parents": [{"id":22, "order":1}]}`,
			},
			args:   []int64{21, 22, 23, 24},
			expect: false,
		},
		{
			desc: "one item skipped",
			items: []string{
				`{"id":21, "type":"Root", "strings":[{"language_id":1, "title":"21"}], "parents": [{"id":21, "order":0}]}`,
				`{"id":22, "type":"Category", "strings":[{"language_id":1, "title":"22"}], "parents": [{"id":21, "order":0}]}`,
				`{"id":23, "type":"Chapter", "strings":[{"language_id":1, "title":"23"}], "parents": [{"id":22, "order":1}]}`,
				`{"id":24, "type":"Chapter", "strings":[{"language_id":1, "title":"23"}], "parents": [{"id":22, "order":1}]}`,
			},
			args:   []int64{21, 22, 23, 24},
			expect: false,
		},
	}

	for _, tCase := range testCases {
		t.Run(tCase.desc, func(t *testing.T) {
			restoreDB(t)
			for _, req := range tCase.items {
				r := &NewItemRequest{}
				err = json.Unmarshal([]byte(req), r)
				if err != nil {
					t.Fatal(err)
				}
				err = srv.insertItem(r)
				if err != nil {
					t.Fatal(err)
				}
			}

			isValidHierarchy, err := srv.Store.Items().IsValidHierarchy(tCase.args)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(tCase.expect, isValidHierarchy)
		})
	}
}

func restoreDB(t *testing.T) {
	cmd := exec.Command("/bin/bash", "testdata/restoredb.sh")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%s", output)
}
