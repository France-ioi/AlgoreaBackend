package database

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTime_ScanString(t *testing.T) {
	tests := []struct {
		name     string
		str      string
		wantErr  bool
		wantTime time.Time
	}{
		{name: "normal datetime", str: "2019-05-30 11:30:12",
			wantTime: time.Date(2019, 5, 30, 11, 30, 12, 0, time.UTC)},
		{name: "zero datetime", str: "0000-00-00 00:00:00", wantTime: time.Time{}},
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
	assert.NoError(t, err)
	assert.Equal(t, tm, value)
}

func TestTime_MarshalJSON(t *testing.T) {
	tm := Time(time.Date(2019, 5, 30, 11, 30, 15, 0, time.UTC))
	result, err := (&tm).MarshalJSON()
	assert.NoError(t, err)
	assert.Equal(t, []byte(`"2019-05-30T11:30:15Z"`), result)
}
