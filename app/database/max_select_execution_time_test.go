package database

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database/mysqldb"
	"github.com/France-ioi/AlgoreaBackend/v2/app/logging"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers/testoutput"
)

func TestInjectMaxExecutionTimeHint(t *testing.T) {
	const hint500 = " /*+ MAX_EXECUTION_TIME(500) */"
	tests := []struct {
		name     string
		query    string
		duration time.Duration
		want     string
	}{
		{
			name:     "plain SELECT",
			query:    "SELECT 1",
			duration: 500 * time.Millisecond,
			want:     "SELECT" + hint500 + " 1",
		},
		{
			name:     "leading whitespace",
			query:    "  \n\tSELECT id FROM t",
			duration: 500 * time.Millisecond,
			want:     "  \n\tSELECT" + hint500 + " id FROM t",
		},
		{
			name:     "leading line comment",
			query:    "-- comment\nSELECT 1",
			duration: 500 * time.Millisecond,
			want:     "-- comment\nSELECT" + hint500 + " 1",
		},
		{
			name:     "leading block comment",
			query:    "/* block */ SELECT 1",
			duration: 500 * time.Millisecond,
			want:     "/* block */ SELECT" + hint500 + " 1",
		},
		{
			name:     "WITH CTE toQuery shape",
			query:    "WITH `t1` AS (SELECT * FROM `otherTable` ) SELECT * FROM `myTable`",
			duration: 500 * time.Millisecond,
			want:     "WITH `t1` AS (SELECT * FROM `otherTable` ) SELECT" + hint500 + " * FROM `myTable`",
		},
		{
			name:     "multiple CTEs with backtick names",
			query:    "WITH `t1` AS (SELECT * FROM `otherTable` ), `t2` AS (SELECT * FROM `thirdTable` ) SELECT * FROM `myTable`",
			duration: 500 * time.Millisecond,
			want:     "WITH `t1` AS (SELECT * FROM `otherTable` ), `t2` AS (SELECT * FROM `thirdTable` ) SELECT" + hint500 + " * FROM `myTable`",
		},
		{
			name:     "WITH RECURSIVE",
			query:    "WITH RECURSIVE t AS (SELECT 1 AS n UNION SELECT n+1 FROM t WHERE n < 3) SELECT * FROM t",
			duration: 500 * time.Millisecond,
			want:     "WITH RECURSIVE t AS (SELECT 1 AS n UNION SELECT n+1 FROM t WHERE n < 3) SELECT" + hint500 + " * FROM t",
		},
		{
			name:     "CTE with nested SELECT and string containing select",
			query:    "WITH t AS (SELECT 'select' AS s, (SELECT 1) AS n) SELECT * FROM t",
			duration: 500 * time.Millisecond,
			want:     "WITH t AS (SELECT 'select' AS s, (SELECT 1) AS n) SELECT" + hint500 + " * FROM t",
		},
		{
			name:     "parenthesised outer block",
			query:    "(SELECT 1)",
			duration: 500 * time.Millisecond,
			want:     "(SELECT" + hint500 + " 1)",
		},
		{
			name:     "UNION first SELECT",
			query:    "SELECT 1 UNION SELECT 2",
			duration: 500 * time.Millisecond,
			want:     "SELECT" + hint500 + " 1 UNION SELECT 2",
		},
		{
			name:     "parenthesised UNION",
			query:    "(SELECT 1) UNION (SELECT 2)",
			duration: 500 * time.Millisecond,
			want:     "(SELECT" + hint500 + " 1) UNION (SELECT 2)",
		},
		{
			name:     "already hinted",
			query:    "SELECT /*+ MAX_EXECUTION_TIME(100) */ 1",
			duration: 500 * time.Millisecond,
			want:     "SELECT /*+ MAX_EXECUTION_TIME(100) */ 1",
		},
		{
			name:     "INSERT unchanged",
			query:    "INSERT INTO t VALUES (1)",
			duration: 500 * time.Millisecond,
			want:     "INSERT INTO t VALUES (1)",
		},
		{
			name:     "UPDATE unchanged",
			query:    "UPDATE t SET a = 1",
			duration: 500 * time.Millisecond,
			want:     "UPDATE t SET a = 1",
		},
		{
			name:     "SET unchanged",
			query:    "SET @a = 1",
			duration: 500 * time.Millisecond,
			want:     "SET @a = 1",
		},
		{
			name:     "zero duration",
			query:    "SELECT 1",
			duration: 0,
			want:     "SELECT 1",
		},
		{
			name:     "negative duration",
			query:    "SELECT 1",
			duration: -time.Second,
			want:     "SELECT 1",
		},
		{
			name:     "sub-millisecond rounds up to 1ms",
			query:    "SELECT 1",
			duration: 100 * time.Microsecond,
			want:     "SELECT /*+ MAX_EXECUTION_TIME(1) */ 1",
		},
		{
			name:     "lowercase select",
			query:    "select id from t",
			duration: 500 * time.Millisecond,
			want:     "select" + hint500 + " id from t",
		},
		{
			name:     "whitespace and comments only",
			query:    "  /* c */ -- x\n",
			duration: 500 * time.Millisecond,
			want:     "  /* c */ -- x\n",
		},
		{
			name:     "WITH without outer SELECT",
			query:    "WITH t AS (SELECT 1 AS a)",
			duration: 500 * time.Millisecond,
			want:     "WITH t AS (SELECT 1 AS a)",
		},
		{
			name:     "WITH body line and hash comments",
			query:    "WITH t AS (SELECT 1) -- c1\n# c2\n SELECT * FROM t",
			duration: 500 * time.Millisecond,
			want:     "WITH t AS (SELECT 1) -- c1\n# c2\n SELECT" + hint500 + " * FROM t",
		},
		{
			name:     "WITH body block comment",
			query:    "WITH t AS (SELECT 1) /* mid */ SELECT * FROM t",
			duration: 500 * time.Millisecond,
			want:     "WITH t AS (SELECT 1) /* mid */ SELECT" + hint500 + " * FROM t",
		},
		{
			name:     "leading hash comment",
			query:    "# hi\nSELECT 1",
			duration: 500 * time.Millisecond,
			want:     "# hi\nSELECT" + hint500 + " 1",
		},
		{
			name:     "dash not a comment inside WITH scan",
			query:    "WITH t AS (SELECT 1- -1 AS n) SELECT * FROM t",
			duration: 500 * time.Millisecond,
			want:     "WITH t AS (SELECT 1- -1 AS n) SELECT" + hint500 + " * FROM t",
		},
		{
			name:     "slash not a comment inside WITH scan",
			query:    "WITH t AS (SELECT 1/2 AS n) SELECT * FROM t",
			duration: 500 * time.Millisecond,
			want:     "WITH t AS (SELECT 1/2 AS n) SELECT" + hint500 + " * FROM t",
		},
		{
			name:     "string with backslash and doubled quotes",
			query:    `WITH t AS (SELECT 'a\'b' AS s, "c""d" AS s2) SELECT * FROM t`,
			duration: 500 * time.Millisecond,
			want:     `WITH t AS (SELECT 'a\'b' AS s, "c""d" AS s2) SELECT` + hint500 + ` * FROM t`,
		},
		{
			name:     "doubled backticks in CTE name",
			query:    "WITH `t``x` AS (SELECT 1 AS a) SELECT * FROM `t``x`",
			duration: 500 * time.Millisecond,
			want:     "WITH `t``x` AS (SELECT 1 AS a) SELECT" + hint500 + " * FROM `t``x`",
		},
		{
			name:     "unclosed string returns unchanged when no top-level SELECT found after",
			query:    "WITH t AS (SELECT 'oops) ",
			duration: 500 * time.Millisecond,
			want:     "WITH t AS (SELECT 'oops) ",
		},
		{
			name:     "unclosed backtick ident",
			query:    "WITH `t AS (SELECT 1) SELECT 2",
			duration: 500 * time.Millisecond,
			want:     "WITH `t AS (SELECT 1) SELECT 2",
		},
		{
			name:     "unclosed block comment at start",
			query:    "/* unterminated SELECT 1",
			duration: 500 * time.Millisecond,
			want:     "/* unterminated SELECT 1",
		},
		{
			name:     "line comment at EOF",
			query:    "SELECT 1 --",
			duration: 500 * time.Millisecond,
			want:     "SELECT" + hint500 + " 1 --",
		},
		{
			name:     "keyword prefix without match",
			query:    "SELECTION 1",
			duration: 500 * time.Millisecond,
			want:     "SELECTION 1",
		},
		{
			name:     "identifier before WITH",
			query:    "XWITH t AS (SELECT 1) SELECT 2",
			duration: 500 * time.Millisecond,
			want:     "XWITH t AS (SELECT 1) SELECT 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, injectMaxExecutionTimeHint(tt.query, tt.duration))
		})
	}
}

func TestContextWithMaxSelectExecutionTime(t *testing.T) {
	ctx := ContextWithMaxSelectExecutionTime(context.Background(), 2*time.Second)
	assert.Equal(t, 2*time.Second, MaxSelectExecutionTimeFromContext(ctx))
	assert.Equal(t, time.Duration(0), MaxSelectExecutionTimeFromContext(context.Background()))
}

func TestIsQueryTimeoutError(t *testing.T) {
	assert.True(t, IsQueryTimeoutError(&mysql.MySQLError{
		Number:  uint16(mysqldb.QueryTimeoutError),
		Message: "Query execution was interrupted, maximum statement execution time exceeded",
	}))
	assert.False(t, IsQueryTimeoutError(&mysql.MySQLError{Number: uint16(mysqldb.DeadlockError)}))
	assert.False(t, IsQueryTimeoutError(assert.AnError))
}

func TestQueryTimeoutError_IsNotRetryable(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	err := &mysql.MySQLError{Number: uint16(mysqldb.QueryTimeoutError)}
	assert.False(t, isRetryableError(err))

	ctx, _, _ := logging.NewContextWithNewMockLogger()
	calls := 0
	got := retryOnRetriableError(ctx, func() error {
		calls++
		return err
	})
	require.Equal(t, err, got)
	assert.Equal(t, 1, calls)
}

func TestSQLDBWrapper_QueryRow_InjectsMaxExecutionTimeHint(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	ctx, _, _ := logging.NewContextWithNewMockLogger()
	ctx = ContextWithMaxSelectExecutionTime(ctx, 500*time.Millisecond)
	db, mock := NewDBMock(ctx)
	defer func() { _ = db.Close() }()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT /*+ MAX_EXECUTION_TIME(500) */ 1")).
		WillReturnRows(mock.NewRows([]string{"1"}).AddRow(1))

	dbWrapper, ok := db.db.CommonDB().(*sqlDBWrapper)
	require.True(t, ok)
	var value int
	require.NoError(t, dbWrapper.QueryRow("SELECT 1").Scan(&value))
	assert.Equal(t, 1, value)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHasSQLKeywordAt_EdgeCases(t *testing.T) {
	assert.False(t, hasSQLKeywordAt("SEL", 0, "SELECT"))
	assert.False(t, hasSQLKeywordAt("xSELECT", 1, "SELECT")) // letter before
	assert.True(t, hasSQLKeywordAt("(SELECT", 1, "SELECT"))
}

func TestIsSQLLineCommentStart_AtEOF(t *testing.T) {
	assert.True(t, isSQLLineCommentStart("--", 0))
	assert.True(t, isSQLLineCommentStart("-- ", 0))
	assert.False(t, isSQLLineCommentStart("---", 0))
}
