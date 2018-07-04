// create_by : lengrongfu
package gdbcerrors

import (
	"fmt"
	"errors"
)

// gloab error
var (
	ErrDsn = errors.New("invalid dsn, dsnName : {user}:{password}@{addr}/{dbName}?paramsKey1={paramsValue1}&paramsKey2={paramsValues2}")
	ErrNoTLS = errors.New("mysql service does not support TLS")
	ErrPackentLength = errors.New("more than max packet length 2^24-1")
	ErrAuthPluginNoSupport = errors.New("mysql service no support this auth plugin")
	ErrNoSecureAuthPlugin = errors.New("this auth plugin is no secure! place user secure auth plugin; Example: mysql_native_password")
	ErrPassowrdNoEmpty = errors.New("auth plugin is sha256 password is not empty")
	ErrEOFPacket = errors.New("this packect is EOF")
)


// GdbcError 自定义error
type GdbcError struct{
	Code    uint16
	Message string
	State   string
}

func(g *GdbcError) Error() string {
	return fmt.Sprintf("GdbcError code is %d, error message is %s , state is %s", g.Code, g.State, g.Message)
}

