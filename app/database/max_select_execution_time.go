package database

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

const sqlBlockCommentMarkerLen = 2 // length of "/*" and "*/"

// ContextWithMaxSelectExecutionTime attaches a wall-clock cap for read-only SELECTs.
//
// Enforced server-side via the MAX_EXECUTION_TIME optimizer hint rather than a context deadline,
// so that the cap still applies when the Lambda is killed, and so that it cannot touch writes
// (MySQL applies max_execution_time to read-only SELECTs only).
func ContextWithMaxSelectExecutionTime(ctx context.Context, d time.Duration) context.Context {
	return context.WithValue(ctx, maxSelectExecutionTimeContextKey, d)
}

// MaxSelectExecutionTimeFromContext returns the SELECT wall-clock cap attached to ctx, or 0 if none.
func MaxSelectExecutionTimeFromContext(ctx context.Context) time.Duration {
	return maxSelectExecutionTimeFromContext(ctx)
}

func maxSelectExecutionTimeFromContext(ctx context.Context) time.Duration {
	d, _ := ctx.Value(maxSelectExecutionTimeContextKey).(time.Duration)
	return d
}

// injectMaxExecutionTimeHint inserts the hint after the top-level SELECT, or returns query
// unchanged if there is no top-level SELECT to attach it to.
func injectMaxExecutionTimeHint(query string, d time.Duration) string {
	if d <= 0 {
		return query
	}

	execBudgetMillis := int64((d + time.Millisecond - 1) / time.Millisecond)

	selectPos := locateTopLevelSelectPosition(query)
	if selectPos < 0 {
		return query
	}

	afterSelect := selectPos + len("SELECT")
	rest := skipSQLWhitespace(query, afterSelect)
	if rest < len(query) && strings.HasPrefix(query[rest:], "/*+") {
		return query
	}

	hint := fmt.Sprintf(" /*+ MAX_EXECUTION_TIME(%d) */", execBudgetMillis)
	return query[:afterSelect] + hint + query[afterSelect:]
}

func locateTopLevelSelectPosition(query string) int {
	pos := skipLeadingParensAndWhitespace(query)
	if pos >= len(query) {
		return -1
	}
	if hasSQLKeywordAt(query, pos, "WITH") {
		return locateSelectAfterWithClause(query, pos)
	}
	if hasSQLKeywordAt(query, pos, "SELECT") {
		return pos
	}
	return -1
}

func skipLeadingParensAndWhitespace(query string) int {
	pos := skipSQLWhitespaceAndComments(query, 0)
	for pos < len(query) {
		pos = skipSQLWhitespaceAndComments(query, pos)
		if pos < len(query) && query[pos] == '(' {
			pos++
			continue
		}
		break
	}
	return skipSQLWhitespaceAndComments(query, pos)
}

func locateSelectAfterWithClause(query string, withPos int) int {
	pos := withPos + len("WITH")
	pos = skipSQLWhitespaceAndComments(query, pos)
	if hasSQLKeywordAt(query, pos, "RECURSIVE") {
		pos += len("RECURSIVE")
		pos = skipSQLWhitespaceAndComments(query, pos)
	}
	return findTopLevelSelect(query, pos)
}

type topLevelSelectScanner struct {
	query string
	pos   int
	depth int
}

func findTopLevelSelect(query string, start int) int {
	scanner := topLevelSelectScanner{query: query, pos: start}
	for scanner.pos < len(scanner.query) {
		if selectPos, found := scanner.step(); found {
			return selectPos
		}
	}
	return -1
}

func (scanner *topLevelSelectScanner) step() (selectPos int, found bool) {
	switch scanner.query[scanner.pos] {
	case '(':
		scanner.depth++
		scanner.pos++
	case ')':
		if scanner.depth > 0 {
			scanner.depth--
		}
		scanner.pos++
	case '\'', '"':
		scanner.pos = skipSQLQuotedString(scanner.query, scanner.pos)
	case '`':
		scanner.pos = skipSQLQuotedIdent(scanner.query, scanner.pos)
	case '-':
		return scanner.stepFromDash()
	case '#':
		scanner.pos = skipSQLLineComment(scanner.query, scanner.pos)
	case '/':
		return scanner.stepFromSlash()
	default:
		return scanner.stepDefault()
	}
	return -1, false
}

func (scanner *topLevelSelectScanner) stepFromDash() (int, bool) {
	if scanner.pos+1 < len(scanner.query) && scanner.query[scanner.pos+1] == '-' &&
		isSQLLineCommentStart(scanner.query, scanner.pos) {
		scanner.pos = skipSQLLineComment(scanner.query, scanner.pos)
		return -1, false
	}
	scanner.pos++
	return -1, false
}

func (scanner *topLevelSelectScanner) stepFromSlash() (int, bool) {
	if scanner.pos+1 < len(scanner.query) && scanner.query[scanner.pos+1] == '*' {
		scanner.pos = skipSQLBlockComment(scanner.query, scanner.pos)
		return -1, false
	}
	scanner.pos++
	return -1, false
}

func (scanner *topLevelSelectScanner) stepDefault() (int, bool) {
	if scanner.depth == 0 && hasSQLKeywordAt(scanner.query, scanner.pos, "SELECT") {
		return scanner.pos, true
	}
	_, size := utf8.DecodeRuneInString(scanner.query[scanner.pos:])
	scanner.pos += size
	return -1, false
}

func hasSQLKeywordAt(query string, pos int, keyword string) bool {
	if pos+len(keyword) > len(query) {
		return false
	}
	if !strings.EqualFold(query[pos:pos+len(keyword)], keyword) {
		return false
	}
	if pos > 0 {
		r, _ := utf8.DecodeLastRuneInString(query[:pos])
		if isSQLIdentRune(r) {
			return false
		}
	}
	end := pos + len(keyword)
	if end < len(query) {
		r, _ := utf8.DecodeRuneInString(query[end:])
		if isSQLIdentRune(r) {
			return false
		}
	}
	return true
}

func isSQLIdentRune(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '$'
}

func skipSQLWhitespace(query string, pos int) int {
	for pos < len(query) {
		r, size := utf8.DecodeRuneInString(query[pos:])
		if !unicode.IsSpace(r) {
			break
		}
		pos += size
	}
	return pos
}

func skipSQLWhitespaceAndComments(query string, pos int) int {
	for pos < len(query) {
		pos = skipSQLWhitespace(query, pos)
		if pos >= len(query) {
			break
		}
		newPos, consumed := skipLeadingCommentAt(query, pos)
		if !consumed {
			break
		}
		pos = newPos
	}
	return pos
}

func skipLeadingCommentAt(query string, pos int) (int, bool) {
	if query[pos] == '#' {
		return skipSQLLineComment(query, pos), true
	}
	if pos+1 < len(query) && query[pos] == '-' && query[pos+1] == '-' && isSQLLineCommentStart(query, pos) {
		return skipSQLLineComment(query, pos), true
	}
	if pos+1 < len(query) && query[pos] == '/' && query[pos+1] == '*' {
		return skipSQLBlockComment(query, pos), true
	}
	return pos, false
}

// MySQL treats "--" as a line comment only when followed by whitespace or end-of-input.
func isSQLLineCommentStart(query string, pos int) bool {
	if pos+sqlBlockCommentMarkerLen >= len(query) {
		return true
	}
	r, _ := utf8.DecodeRuneInString(query[pos+sqlBlockCommentMarkerLen:])
	return unicode.IsSpace(r)
}

func skipSQLLineComment(query string, pos int) int {
	for pos < len(query) && query[pos] != '\n' && query[pos] != '\r' {
		pos++
	}
	return pos
}

func skipSQLBlockComment(query string, pos int) int {
	pos += sqlBlockCommentMarkerLen // skip /*
	for pos+1 < len(query) {
		if query[pos] == '*' && query[pos+1] == '/' {
			return pos + sqlBlockCommentMarkerLen
		}
		pos++
	}
	return len(query)
}

func skipSQLQuotedString(query string, pos int) int {
	quote := query[pos]
	pos++
	for pos < len(query) {
		c := query[pos]
		if c == '\\' {
			pos += sqlBlockCommentMarkerLen
			continue
		}
		if c == quote {
			// MySQL doubled-quote escape: '' or ""
			if pos+1 < len(query) && query[pos+1] == quote {
				pos += sqlBlockCommentMarkerLen
				continue
			}
			return pos + 1
		}
		pos++
	}
	return len(query)
}

func skipSQLQuotedIdent(query string, pos int) int {
	pos++ // skip opening `
	for pos < len(query) {
		if query[pos] == '`' {
			if pos+1 < len(query) && query[pos+1] == '`' {
				pos += sqlBlockCommentMarkerLen
				continue
			}
			return pos + 1
		}
		pos++
	}
	return len(query)
}
