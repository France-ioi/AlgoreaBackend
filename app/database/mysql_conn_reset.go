package database

import (
	"context"
	"database/sql/driver"
	"net"
	"sync/atomic"
	"time"
	_ "unsafe" // for go:linkname

	"github.com/go-sql-driver/mysql"

	log "github.com/France-ioi/AlgoreaBackend/v2/app/logging"
)

// A buffer which is used for both reading and writing.
// This is possible since communication on each connection is synchronous.
// In other words, we can't write and read simultaneously on the same connection.
// The buffer is similar to bufio.Reader / Writer but zero-copy-ish
// Also highly optimized for this particular use case.
// This buffer is backed by two byte slices in a double-buffering scheme.
type buffer struct {
	_ []byte // buf is a byte buffer who's length and capacity are equal.
	_ net.Conn
	_ int
	_ int
	_ time.Duration
	_ [2][]byte // dbuf is an array with the two byte slices that back this buffer
	_ uint      // flipccnt is the current buffer counter for double-buffering
}

type mysqlResult struct {
	// One entry in both slices is created for every executed statement result.
	_ []int64
	_ []int64
}

type connector struct {
	_ *mysql.Config // immutable private copy.
	_ string        // Encoded connection attributes.
}

// https://dev.mysql.com/doc/internals/en/capability-flags.html#packet-Protocol::CapabilityFlags
type clientFlag uint32

// http://dev.mysql.com/doc/internals/en/status-flags.html
type statusFlag uint16

// noCopy may be embedded into structs which must not be copied
// after the first use.
//
// See https://github.com/golang/go/issues/8005#issuecomment-190753527
// for details.
type (
	noCopy     struct{}
	atomicBool = atomic.Bool
)

// atomicError is a wrapper for atomically accessed error values.
type atomicError struct {
	_ noCopy
	_ atomic.Value
}

type mysqlConn struct {
	_ buffer
	_ net.Conn
	_ net.Conn    // underlying connection when netConn is TLS connection.
	_ mysqlResult // managed by clearResult() and handleOkPacket().
	_ *mysql.Config
	_ *connector
	_ int
	_ int
	_ time.Duration
	_ clientFlag
	_ statusFlag
	_ uint8
	_ bool

	// for context support (Go 1.8+)
	_      bool
	_      chan<- context.Context
	_      chan struct{}
	_      chan<- struct{}
	_      atomicError // set non-nil if conn is canceled
	closed atomicBool  // set when conn is closed, before closech is closed
}

const comResetConnection byte = 31

// Reset resets the MySQL connection.
// The implementation is based on the original Ping method from the go-sql-driver/mysql package,
// but it sends a comResetConnection command instead of a comPing command.
// Also, it logs the command if raw SQL queries logging is enabled.
func (mc *mysqlConn) Reset(ctx context.Context) (err error) {
	if log.SharedLogger.IsRawSQLQueriesLoggingEnabled() {
		startTime := time.Now()
		defer func() {
			log.SharedLogger.WithContext(ctx).WithFields(map[string]interface{}{
				"type": "db", "err": err, "duration": time.Since(startTime).String(),
			}).Info("sql-conn-reset")
		}()
	}

	if mc.closed.Load() {
		mysqlConnLog(mc, mysql.ErrInvalidConn)
		return driver.ErrBadConn
	}

	if err = mysqlConnWatchCancel(mc, ctx); err != nil {
		return
	}
	defer mysqlConnFinish(mc)

	handleOk := mysqlConnClearResult(mc)

	if err = mysqlConnWriteCommandPacket(mc, comResetConnection); err != nil {
		return mysqlConnMarkBadConn(mc, err)
	}

	return mysqlOKHandlerReadResultOK(handleOk)
}

//go:linkname mysqlConnLog github.com/go-sql-driver/mysql.(*mysqlConn).log
func mysqlConnLog(*mysqlConn, ...any)

//go:linkname mysqlConnWatchCancel github.com/go-sql-driver/mysql.(*mysqlConn).watchCancel
func mysqlConnWatchCancel(*mysqlConn, context.Context) error //nolint: revive

//go:linkname mysqlConnFinish github.com/go-sql-driver/mysql.(*mysqlConn).finish
func mysqlConnFinish(*mysqlConn)

type okHandler mysqlConn

//go:linkname mysqlConnClearResult github.com/go-sql-driver/mysql.(*mysqlConn).clearResult
func mysqlConnClearResult(*mysqlConn) *okHandler

//go:linkname mysqlConnWriteCommandPacket github.com/go-sql-driver/mysql.(*mysqlConn).writeCommandPacket
func mysqlConnWriteCommandPacket(*mysqlConn, byte) error

//go:linkname mysqlConnMarkBadConn github.com/go-sql-driver/mysql.(*mysqlConn).markBadConn
func mysqlConnMarkBadConn(*mysqlConn, error) error

//go:linkname mysqlOKHandlerReadResultOK github.com/go-sql-driver/mysql.(*okHandler).readResultOK
func mysqlOKHandlerReadResultOK(*okHandler) error
