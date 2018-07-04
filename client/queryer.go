package client

import (
	"database/sql/driver"
	"gdbc/constant"
	"gdbc/gdbcerrors"
	"gdbc/utils"
	"github.com/golang/glog"
	"strings"
)

//Query implemented by a Conn
func (q *Conn) Query(query string, args []driver.Value) (driver.Rows, error) {
	if q.netConn == nil {
		return nil, driver.ErrBadConn
	}
	argNum := strings.Count(query, "?")
	if argNum != len(args) {
		e := new(gdbcerrors.GdbcError)
		e.Message = "传入参数和查询参数不一致," + "查询语句为:" + query
		return nil, e
	}
	if len(args)+4 > constant.MaxPayloadLength {
		return nil, driver.ErrSkip
	}
	if len(args) != 0 {
		query = q.ParseParam(query, args)
	}
	//打印sql
	if q.cnf.DEBUG {
		glog.Infof("sql:[%s]", string(query))
	}
	comd := constant.ComQuery
	comds := []byte{comd}
	comds = append(comds, []byte(query)...)
	if err := q.WriteCommand(comds); err != nil {
		glog.Error(err)
		return nil, err
	}
	rows := new(GdbcRows)
	rows.c = q
	rows.protocolType = constant.TextProtocol
	if err := rows.HandlerRows(); err != nil {
		glog.Error(err)
		return nil, err
	}
	return rows, nil
}

//ParseParam 解析请求参数
func (q *Conn) ParseParam(query string, args []driver.Value) string {
	var param []byte
	for k, v := range args {
		_, param = utils.DriverValueHandler(v, constant.TextProtocol)
		query = strings.Replace(query, "?", string(param), k+1)
	}
	return query
}
