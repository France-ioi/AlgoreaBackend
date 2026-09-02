package database

import (
	"context"
	"database/sql/driver"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"
)

//nolint:gochecknoglobals // session settings are re-applied by the shared driver wrapper after every COM_RESET_CONNECTION
var (
	sessionParamsMu sync.RWMutex
	sessionParams   map[string]string
)

var (
	sessionParamNameRegexp  = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)
	sessionParamValueRegexp = regexp.MustCompile(`^(?:[A-Za-z0-9_.+-]+|'[^'\\]*')$`)
)

// SetSessionParams configures MySQL session variables re-applied after every connection reset.
// Pass nil or an empty map to clear them. Names and values are validated before they are stored.
func SetSessionParams(params map[string]string) error {
	if err := validateSessionParams(params); err != nil {
		return err
	}
	sessionParamsMu.Lock()
	defer sessionParamsMu.Unlock()
	sessionParams = cloneStringMap(params)
	return nil
}

func getSessionParams() map[string]string {
	sessionParamsMu.RLock()
	defer sessionParamsMu.RUnlock()
	return cloneStringMap(sessionParams)
}

func cloneStringMap(m map[string]string) map[string]string {
	if len(m) == 0 {
		return nil
	}
	cloned := make(map[string]string, len(m))
	for k, v := range m {
		cloned[k] = v
	}
	return cloned
}

func validateSessionParams(params map[string]string) error {
	for name, value := range params {
		if !sessionParamNameRegexp.MatchString(name) {
			return fmt.Errorf("invalid database session param name %q", name)
		}
		if !sessionParamValueRegexp.MatchString(value) {
			return fmt.Errorf("invalid database session param value for %q", name)
		}
	}
	return nil
}

func (conn *mysqlConnWrapper) applySessionParams(ctx context.Context) error {
	if len(conn.sessionParams) == 0 {
		return nil
	}

	// Re-validates params cloned from getSessionParams(); kept as a guard at the SQL concatenation site
	// and for tests that construct mysqlConnWrapper directly.
	if err := validateSessionParams(conn.sessionParams); err != nil {
		return err
	}

	names := make([]string, 0, len(conn.sessionParams))
	for name := range conn.sessionParams {
		names = append(names, name)
	}
	sort.Strings(names)

	// Build a single SET without placeholders: MySQL rejects prepared statements for some
	// session system variables. Names/values are validated; string vars include quotes in the value
	// (same convention as go-sql-driver DSN params).
	assignments := make([]string, len(names))
	for i, name := range names {
		assignments[i] = "@@SESSION." + name + "=" + conn.sessionParams[name]
	}

	//nolint:forcetypeassert // panic if conn.conn does not implement driver.ExecerContext
	_, err := conn.conn.(driver.ExecerContext).ExecContext(ctx, "SET "+strings.Join(assignments, ", "), nil)
	return err
}
