package logging

import (
	"database/sql/driver"
	"testing"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
)

type timeType time.Time

// Value returns a timeType Value (*time.Time).
func (t *timeType) Value() (driver.Value, error) {
	if t == nil {
		return nil, nil
	}
	return (*time.Time)(t).UTC().Format("2006-01-02 15:04:05.999999"), nil
}

func Test_GormLogFormatter(t *testing.T) {
	timeValue := time.Date(2022, 8, 2, 15, 4, 5, 123456789, time.UTC)
	timeValuePtr := timeValue
	timeTypeValue := timeType(timeValue)
	timeTypeValuePtr := &timeTypeValue
	int64Value := int64(12345)
	int64Ptr := &int64Value
	nilValue := (*int64)(nil)
	nilValuePtr := &nilValue

	tests := []struct {
		name     string
		args     []interface{}
		expected []interface{}
	}{
		{
			name: "scalar numbers and booleans",
			args: []interface{}{
				"sql", "file:line", time.Duration(52739320000),
				"INSERT INTO `t` (`c1`, `c2`, `c3`, `c4`, `c5`, `c6`, `c7`, `c8`, `c9`, `c10`, `c11`, `c12`, `c13`) VALUES " +
					"(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
				[]interface{}{
					1, int8(2), int16(3), int32(4), int64(5), uint(6), uint8(7), uint16(8), uint32(9), uint64(10),
					float32(0.11), float64(0.12), true,
				},
				int64(1),
			},
			expected: []interface{}{
				"\033[35m(file:line)\033[0m",
				"\n\033[33m[2021-08-02 15:04:05]\033[0m",
				" \033[36;1m[52739.32ms]\033[0m ",
				"INSERT INTO `t` (`c1`, `c2`, `c3`, `c4`, `c5`, `c6`, `c7`, `c8`, `c9`, `c10`, `c11`, `c12`, `c13`) VALUES " +
					"(1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 0.11, 0.12, true)",
				" \n\x1b[36;31m[1 rows affected ]\x1b[0m ",
			},
		},
		{
			name: "time values",
			args: []interface{}{
				"sql", "file:line", time.Duration(52739320000),
				"INSERT INTO `t` (`c1`, `c2`, `c3`, `c4`, `c5`, `c6`) VALUES (?, ?, ?, ?, ?, ?)",
				[]interface{}{timeValue, timeValuePtr, timeTypeValue, timeTypeValuePtr, (*timeType)(nil), time.Time{}},
				int64(1),
			},
			expected: []interface{}{
				"\033[35m(file:line)\033[0m",
				"\n\033[33m[2021-08-02 15:04:05]\033[0m",
				" \033[36;1m[52739.32ms]\033[0m ",
				"INSERT INTO `t` (`c1`, `c2`, `c3`, `c4`, `c5`, `c6`) VALUES " +
					"('2022-08-02 15:04:05.123456789', '2022-08-02 15:04:05.123456789', " +
					"'2022-08-02 15:04:05.123456', '2022-08-02 15:04:05.123456', NULL, '0000-00-00 00:00:00')",
				" \n\x1b[36;31m[1 rows affected ]\x1b[0m ",
			},
		},
		{
			name: "$n placeholders",
			args: []interface{}{
				"sql", "file:line", time.Duration(52739320000),
				"INSERT INTO `t` (`c1`, `c2`, `c3`, `c4`) VALUES ($1, $2, $3, $3)",
				[]interface{}{int64(123451234512345), "testsessiontestsessiontestsessio", int64(1), int64(2)},
				int64(1),
			},
			expected: []interface{}{
				"\033[35m(file:line)\033[0m",
				"\n\033[33m[2021-08-02 15:04:05]\033[0m",
				" \033[36;1m[52739.32ms]\033[0m ",
				"INSERT INTO `t` (`c1`, `c2`, `c3`, `c4`) VALUES (123451234512345, 'testsessiontestsessiontestsessio', 1, 1)",
				" \n\x1b[36;31m[1 rows affected ]\x1b[0m ",
			},
		},
		{
			name: "bytes values",
			args: []interface{}{
				"sql", "file:line", time.Duration(52739320000),
				"INSERT INTO `t` (`c1`, `c2`) VALUES ($1, $2)",
				[]interface{}{[]byte("test"), []byte{0x01, 0x02, 0x03}},
				int64(1),
			},
			expected: []interface{}{
				"\033[35m(file:line)\033[0m",
				"\n\033[33m[2021-08-02 15:04:05]\033[0m",
				" \033[36;1m[52739.32ms]\033[0m ",
				"INSERT INTO `t` (`c1`, `c2`) VALUES ('test', '<binary>')",
				" \n\x1b[36;31m[1 rows affected ]\x1b[0m ",
			},
		},
		{
			name: "pointer values",
			args: []interface{}{
				"sql", "file:line", time.Duration(52739320000),
				"INSERT INTO `t` (`c1`, `c2`) VALUES (?, ?)",
				[]interface{}{int64Ptr, nilValuePtr},
				int64(1),
			},
			expected: []interface{}{
				"\033[35m(file:line)\033[0m",
				"\n\033[33m[2021-08-02 15:04:05]\033[0m",
				" \033[36;1m[52739.32ms]\033[0m ",
				"INSERT INTO `t` (`c1`, `c2`) VALUES (12345, NULL)",
				" \n\x1b[36;31m[1 rows affected ]\x1b[0m ",
			},
		},
		{
			name: "nil value",
			args: []interface{}{
				"sql", "file:line", time.Duration(52739320000),
				"INSERT INTO `t` (`c1`) VALUES (?)",
				[]interface{}{nil},
				int64(1),
			},
			expected: []interface{}{
				"\033[35m(file:line)\033[0m",
				"\n\033[33m[2021-08-02 15:04:05]\033[0m",
				" \033[36;1m[52739.32ms]\033[0m ",
				"INSERT INTO `t` (`c1`) VALUES (NULL)",
				" \n\x1b[36;31m[1 rows affected ]\x1b[0m ",
			},
		},
		{
			name:     "one argument",
			args:     []interface{}{"sql"},
			expected: nil,
		},
		{
			name:     "two arguments",
			args:     []interface{}{"sql", "file:line"},
			expected: []interface{}{"\x1b[33m[2021-08-02 15:04:05]\x1b[0m", "\x1b[35mfile:line\x1b[0m"},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			oldGormNowFunc := gorm.NowFunc
			gorm.NowFunc = func() time.Time { return time.Date(2021, 8, 2, 15, 4, 5, 123456789, time.UTC) }
			defer func() { gorm.NowFunc = oldGormNowFunc }()

			assert.Equal(t, test.expected, gorm.LogFormatter(test.args...))
		})
	}
}
