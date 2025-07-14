package database

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTime_ScanString(t *testing.T) {
	tests := []struct {
		name     string
		str      string
		wantErr  bool
		wantTime time.Time
	}{
		{
			name: "normal datetime", str: "2019-05-30 11:30:12",
			wantTime: time.Date(2019, 5, 30, 11, 30, 12, 0, time.UTC),
		},
		{name: "zero datetime", str: "0000-00-00 00:00:00", wantTime: time.Time{}},
		{
			name: "normal date", str: "2019-05-30",
			wantTime: time.Date(2019, 5, 30, 0, 0, 0, 0, time.UTC),
		},
		{name: "zero date", str: "0000-00-00", wantTime: time.Time{}},
		{
			name: "normal datetime with ms (1 digit)", str: "2019-05-30 11:30:12.1",
			wantTime: time.Date(2019, 5, 30, 11, 30, 12, 100000000, time.UTC),
		},
		{name: "zero datetime with ms (1 digit)", str: "0000-00-00 00:00:00.0", wantTime: time.Time{}},
		{
			name: "normal datetime with ms 2", str: "2019-05-30 11:30:12.12",
			wantTime: time.Date(2019, 5, 30, 11, 30, 12, 120000000, time.UTC),
		},
		{name: "zero datetime with ms (2 digits)", str: "0000-00-00 00:00:00.00", wantTime: time.Time{}},
		{
			name: "normal datetime with ms (3 digits)", str: "2019-05-30 11:30:12.123",
			wantTime: time.Date(2019, 5, 30, 11, 30, 12, 123000000, time.UTC),
		},
		{name: "zero datetime with ms  (3 digits)", str: "0000-00-00 00:00:00.000", wantTime: time.Time{}},
		{
			name: "normal datetime with ms (4 digits)", str: "2019-05-30 11:30:12.1234",
			wantTime: time.Date(2019, 5, 30, 11, 30, 12, 123400000, time.UTC),
		},
		{name: "zero datetime with ms  (4 digits)", str: "0000-00-00 00:00:00.0000", wantTime: time.Time{}},
		{
			name: "normal datetime with ms (5 digits)", str: "2019-05-30 11:30:12.12345",
			wantTime: time.Date(2019, 5, 30, 11, 30, 12, 123450000, time.UTC),
		},
		{name: "zero datetime with ms  (5 digits)", str: "0000-00-00 00:00:00.00000", wantTime: time.Time{}},
		{
			name: "normal datetime with ms (6 digits)", str: "2019-05-30 11:30:12.123456",
			wantTime: time.Date(2019, 5, 30, 11, 30, 12, 123456000, time.UTC),
		},
		{name: "zero datetime with ms  (6 digits)", str: "0000-00-00 00:00:00.000000", wantTime: time.Time{}},
		{name: "wrong datetime", str: "12345-67-89 25:60:70", wantErr: true, wantTime: time.Time{}},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tm := &Time{}
			if err := tm.ScanString(tt.str); (err != nil) != tt.wantErr {
				t.Errorf("ScanString() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, tt.wantTime, time.Time(*tm))
		})
	}
}

func TestTime_Value(t *testing.T) {
	tm := &time.Time{}
	value, err := (*Time)(tm).Value()
	require.NoError(t, err)
	assert.Equal(t, "0001-01-01 00:00:00", value)
}

func TestTime_Value_Nil(t *testing.T) {
	tm := (*Time)(nil)
	value, err := tm.Value()
	require.NoError(t, err)
	assert.Nil(t, value)
}

func TestTime_MarshalJSON(t *testing.T) {
	tm := Time(time.Date(2019, 5, 30, 11, 30, 15, 0, time.UTC))
	result, err := (&tm).MarshalJSON()
	require.NoError(t, err)
	assert.Equal(t, []byte(`"2019-05-30T11:30:15Z"`), result)
}
