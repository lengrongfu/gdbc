package client

import (
	"database/sql/driver"
	"errors"
	"gdbc/constant"
	"io"
)

//GdbcRows TextRows
type GdbcRows struct {
	c               *Conn
	txtResultset    *TextResultset
	binaryResultset *BinaryResultset
	protocolType    int
	//迭代的次数
	iterator int64
}

// Columns 需要重写
func (r *GdbcRows) Columns() []string {
	var columns []string
	if r.protocolType == constant.TextProtocol {
		columns = make([]string, r.txtResultset.columnCount)
		for index, field := range r.txtResultset.columnDefinition41Packets {
			columns[index] = field.name
		}
	} else if r.protocolType == constant.BinaryProtocol {
		columns = make([]string, r.binaryResultset.columnCount)
		for index, field := range r.binaryResultset.columnDefinition41Packets {
			columns[index] = field.name
		}
	}
	return columns
}

//Close ...
func (r *GdbcRows) Close() error {
	r.iterator = -1
	return nil
}

// dest 只能在这里被赋值
//Next iterator row
func (r *GdbcRows) Next(dest []driver.Value) error {
	if r.iterator == -1 {
		return io.ErrUnexpectedEOF
	}
	if r.protocolType != constant.TextProtocol && r.protocolType != constant.BinaryProtocol {
		return io.EOF
	}
	if r.protocolType == constant.TextProtocol && r.iterator >= int64(len(r.txtResultset.rows)) {
		return io.EOF
	}
	if r.protocolType == constant.BinaryProtocol && r.iterator >= int64(len(r.binaryResultset.rows)) {
		return io.EOF
	}
	//处理数据
	var data []driver.Value
	if r.protocolType == constant.TextProtocol {
		data = r.txtResultset.ValueCoverage(r.iterator)
	} else if r.protocolType == constant.BinaryProtocol {
		data = r.binaryResultset.ValueCoverage(r.iterator)
	}
	if data != nil {
		for k, v := range data {
			dest[k] = v
		}
	}
	r.iterator++
	return nil
}

//HandlerRows ...
func (r *GdbcRows) HandlerRows() error {
	if r.protocolType == constant.TextProtocol {
		if resultset, err := HandlerResultSet(r.c); err != nil {
			return err
		} else {
			r.txtResultset = resultset
			return nil
		}
	} else if r.protocolType == constant.BinaryProtocol {
		if binaryResultset, err := HandlerBinaryResultSet(r.c); err != nil {
			return err
		} else {
			r.binaryResultset = binaryResultset
			return nil
		}
	}
	return errors.New(" no support protocol ")
}

//MysqlDriverValueHandler 转换mysql的类型
func mysqlDriverValueHandler(arg []byte, field ColumnDef41Packets, index int) (driver.Value, int) {
	isUnsigned := field.flags&constant.UnsignedFlag != 0
	switch field.typ {
	case constant.MysqlTypeNull:
		return nil, 0
	case constant.MysqlTypeTiny:
		if isUnsigned {

		} else {

		}
	default:
		return nil, 0
	}
	return nil, 0
}
