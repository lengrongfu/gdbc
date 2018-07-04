package client

import (
	"database/sql/driver"
	"encoding/binary"
	"errors"
	"gdbc/constant"
	"gdbc/packets"
	"gdbc/utils"
	"github.com/golang/glog"
	"math"
	"strconv"
	"time"
)

// Resultset is text protocol
//Resultset COM_QUERY 查询结果中的一部分
type Resultset struct {
	columnCount               uint64
	columnDefinition41Packets []ColumnDef41Packets
	// values                    [][]driver.Value
}

//TextResultset 是mysql查询按txt协议返回的
type TextResultset struct {
	Resultset
	rows []TextResultsetRow
}

//HandlerResultSet parse payload to ResultSet
func HandlerResultSet(c *Conn) (*TextResultset, error) {
	resultset := new(TextResultset)
	//读第一个数据包
	if err := c.Payload.ReadPayload(c.netConn); err != nil {
		return nil, err
	}
	if errPackt := packets.ErrPacketHandler(c.Payload.Payload, c.Payload.Capability); errPackt.IsErr() {
		return nil, errors.New(errPackt.String())
	}
	resultset.columnCount, _ = utils.LenencIntDecode(c.Payload.Payload[0:])
	//5.7.5 之前是使用EOF来标示一个数据包结束，之后是使用一个OK包来标志所有的包都结束，主要看 CLIENT_DEPRECATE_EOF 这个参数服务器是否支持
	if c.Payload.Capability&constant.ClientProtocol41 > 0 && resultset.columnCount > 0 {
		//读第二个数据包
		columns := make([]ColumnDef41Packets, 0)
		for i := 0; i < int(resultset.columnCount); i++ {
			if err := c.Payload.ReadPayload(c.netConn); err != nil {
				return nil, err
			}
			payload := c.Payload.Payload
			column := HandlerColumnDef41(payload)
			columns = append(columns, column)
		}
		resultset.columnDefinition41Packets = columns
	}

	if err := c.Payload.ReadPayload(c.netConn); err != nil {
		return nil, err
	}
	//如果没有设置CLIENT_DEPRECATE_EOF 这个标志位就使用EOF包做结束标志位
	if c.Payload.Capability&constant.ClientDeprecateEOF == 0 {
		if eofPack := packets.EOFPacketHandler(c.Payload.Payload, c.Payload.Capability); !eofPack.IsEOF() {
			return nil, errors.New(eofPack.String())
		}
	} else {
		if okPack := packets.OkPacketHandler(c.Payload.Payload, c.Payload.Capability); !okPack.IsOk() {
			return nil, errors.New("请检查 mysql server 服务版本")
		}
	}
	rows := make([]TextResultsetRow, 0)
	for {
		if err := c.Payload.ReadPayload(c.netConn); err != nil {
			return nil, err
		}

		if c.Payload.Capability&constant.ClientDeprecateEOF == 0 {
			//如果是以EOF包结束
			if eofPack := packets.EOFPacketHandler(c.Payload.Payload, c.Payload.Capability); eofPack.IsEOF() {
				break
			}
		} else {
			//如果已OK包结束
			if okPack := packets.OkPacketHandler(c.Payload.Payload, c.Payload.Capability); okPack.IsOk() {
				break
			}
		}

		if errPack := packets.ErrPacketHandler(c.Payload.Payload, c.Payload.Capability); errPack.IsErr() {
			return nil, errors.New(errPack.String())
		}
		txtRow := TextResultsetRow{}
		if len(c.Payload.Payload) == 1 && c.Payload.Payload[0] == 0xfb {
			txtRow.isNull = true
		} else {
			txtRow.isNull = false
			txtRow.values = string(c.Payload.Payload)
		}
		rows = append(rows, txtRow)
	}
	resultset.rows = rows
	return resultset, nil
}

//ValueCoverage text protocol 值转换
func (text *TextResultset) ValueCoverage(iterator int64) []driver.Value {
	row := make([]driver.Value, text.columnCount)
	if text.rows[iterator].isNull {
		return row
	}
	var post int
	rows := []byte(text.rows[iterator].values)
	for index, field := range text.Resultset.columnDefinition41Packets {
		value, n, isNull := utils.LenencStringDecode(rows[post:])
		post += n
		if isNull {
			row[index] = nil
			if field.typ == constant.MysqlTypeTimestamp ||
				field.typ == constant.MysqlTypeDate ||
				field.typ == constant.MysqlTypeTime ||
				field.typ == constant.MysqlTypeDatetime {
				layout := "2006-01-02 00:00:00"
				value = "0000-01-01 00:00:00"
				tim, err := time.Parse(layout, value)
				if err != nil {
					glog.Errorf("time parse error:%s,error is:%+v", value, err)
					continue
				}
				row[index] = tim
			}
			continue
		}
		isUnsigned := field.flags&constant.UnsignedFlag != 0
		switch field.typ {
		case constant.MysqlTypeTiny, constant.MysqlTypeShort, constant.MysqlTypeLong,
			constant.MysqlTypeInt24, constant.MysqlTypeLongLong, constant.MysqlTypeYear,
			constant.MysqlTypeBit:
			if isUnsigned {
				row[index], _ = strconv.ParseUint(value, 10, 64)
			} else {
				row[index], _ = strconv.ParseInt(value, 10, 64)
			}
			break
		case constant.MysqlTypeFloat, constant.MysqlTypeDouble:
			row[index], _ = strconv.ParseFloat(value, 64)
		case constant.MysqlTypeDecimal, constant.MysqlTypeNewDecimal,
			constant.MysqlTypeEnum, constant.MysqlTypeSet,
			constant.MysqlTypeVarString, constant.MysqlTypeString,
			constant.MysqlTypeVarchar:
			row[index] = value
		case constant.MysqlTypeTimestamp, constant.MysqlTypeDate, constant.MysqlTypeTime, constant.MysqlTypeDatetime:
			layout := "2006-01-02 00:00:00"
			tim, err := time.Parse(layout, value)
			if err != nil {
				glog.Error("time parse error:%s", value)
				continue
			}
			row[index] = tim
		default:
			row[index] = value
		}
	}
	return row
}

//BinaryResultset Binary Protocol Resultset
type BinaryResultset struct {
	Resultset
	rows   []BinaryResultsetRow
	result *GdbcResult
}

//HandlerBinaryResultSet ...
func HandlerBinaryResultSet(c *Conn) (*BinaryResultset, error) {
	binaryResult := new(BinaryResultset)
	//读第一个数据包
	if err := c.Payload.ReadPayload(c.netConn); err != nil {
		return nil, err
	}
	if okPackt := packets.OkPacketHandler(c.Payload.Payload, c.Payload.Capability); okPackt.IsOk() {
		binaryResult.result = new(GdbcResult)
		binaryResult.result.affectedRows = okPackt.AffectedRows
		binaryResult.result.lastInsertID = okPackt.LastInsertId
		return binaryResult, nil
	}
	if errPackt := packets.ErrPacketHandler(c.Payload.Payload, c.Payload.Capability); errPackt.IsErr() {
		glog.Errorf("COM_STMT_EXECUTE command,service return err packet:%s", errPackt.String())
		return nil, errors.New(errPackt.String())
	}
	binaryResult.columnCount, _ = utils.LenencIntDecode(c.Payload.Payload)
	if columnDefinition, err := HandlerColumnDef41s(c); err != nil {
		return nil, err
	} else {
		binaryResult.columnDefinition41Packets = columnDefinition
	}
	if results, err := HandlerResultSetRowS(c, binaryResult.columnCount); err != nil {
		return nil, err
	} else {
		binaryResult.rows = results
	}
	return binaryResult, nil
}

//https://dev.mysql.com/doc/internals/en/binary-protocol-value.html 数据解析
//https://dev.mysql.com/doc/internals/en/null-bitmap.html
//ValueCoverage binary protocol coverage
func (binaryPro *BinaryResultset) ValueCoverage(iterator int64) []driver.Value {
	row := make([]driver.Value, binaryPro.columnCount)
	if binaryPro.rows[iterator].header != constant.OkHeader {
		glog.Errorf("binary protocol header is not OK_HEADER,is:%d", binaryPro.rows[iterator].header)
		return nil
	}
	rows := binaryPro.rows[iterator].values
	pos := 0
	nullBitmap := binaryPro.rows[iterator].NULLBitmap
	// var n,num uint6
	for index, field := range binaryPro.columnDefinition41Packets {
		if nullBitmap[(index+2)/8]&(1<<(uint(index+2)%8)) > 0 {
			row[index] = nil
			if field.typ == constant.MysqlTypeTimestamp ||
				field.typ == constant.MysqlTypeDate ||
				field.typ == constant.MysqlTypeTime ||
				field.typ == constant.MysqlTypeDatetime {
				layout := "2006-01-02 00:00:00"
				value := "0000-01-01 00:00:00"
				tim, err := time.Parse(layout, value)
				if err != nil {
					glog.Errorf("time parse error:%s,error is:%+v", value, err)
					continue
				}
				row[index] = tim
			}
			continue
		}
		isUnsigned := field.flags&constant.UnsignedFlag != 0
		switch field.typ {
		case constant.MysqlTypeNull:
			row[index] = nil
			continue
		case constant.MysqlTypeString, constant.MysqlTypeVarchar, constant.MysqlTypeVarString, constant.MysqlTypeEnum,
			constant.MysqlTypeSet, constant.MysqlTypeLongBlog, constant.MysqlTypeMediumBlob, constant.MysqlTypeBlob,
			constant.MysqlTypeTinyBlob, constant.MysqlTypeGeometry, constant.MysqlTypeBit, constant.MysqlTypeDecimal,
			constant.MysqlTypeNewDecimal:
			valueStr, n, _ := utils.LenencStringDecode(rows[pos:])
			row[index] = valueStr
			pos += n
			continue
		case constant.MysqlTypeLongLong:
			if isUnsigned {
				row[index] = binary.LittleEndian.Uint64(rows[pos : pos+8])
			} else {
				row[index] = int64(binary.LittleEndian.Uint64(rows[pos : pos+8]))
			}
			pos += 8
			continue
		case constant.MysqlTypeLong, constant.MysqlTypeInt24:
			if isUnsigned {
				row[index] = binary.LittleEndian.Uint32(rows[pos : pos+4])
			} else {
				row[index] = int32(binary.LittleEndian.Uint32(rows[pos : pos+4]))
			}
			pos += 4
			continue
		case constant.MysqlTypeShort, constant.MysqlTypeYear:
			if isUnsigned {
				row[index] = binary.LittleEndian.Uint16(rows[pos : pos+2])
			} else {
				row[index] = int16(binary.LittleEndian.Uint16(rows[pos : pos+2]))
			}
			pos += 2
			continue
		case constant.MysqlTypeTiny:
			if isUnsigned {
				row[index] = rows[pos : pos+1][0]
			} else {
				row[index] = int8(rows[pos : pos+1][0])
			}
			pos++
			continue
		case constant.MysqlTypeDouble:
			row[index] = math.Float64frombits(binary.LittleEndian.Uint64(rows[pos : pos+8]))
			// row[index] = string(rows[pos : pos+8])
			pos += 8
			continue
		case constant.MysqlTypeFloat:
			row[index] = math.Float32frombits(binary.LittleEndian.Uint32(rows[pos : pos+4]))
			pos += 4
			continue
		case constant.MysqlTypeDate, constant.MysqlTypeDatetime, constant.MysqlTypeTimestamp:
			length := rows[pos]
			if length != 0 && length != 4 && length != 7 && length != 11 {
				glog.Error("time date format error : number of bytes following (valid values: 0, 4, 7, 11)")
				row[index] = nil
				continue
			}
			//if uint(binaryPro.rows[iterator].NULLBitmap[uint(index)>>3])&uint(index) > 0 {
			//	row[index] = nil
			//}
			var err error
			var timeDate time.Time
			timeDate, err = utils.FormatBinaryDateTime(int(length), rows[pos:])
			row[index] = timeDate
			if err != nil {
				glog.Error(err)
				return nil
			}
			pos += int(length)
			continue
		case constant.MysqlTypeTime:
			length := rows[pos]
			if length != 0 && length != 8 && length != 12 {
				glog.Error("MysqlTypeTime format error : number of bytes following (valid values: 0, 8, 12)")
				row[index] = nil
				continue
			}
			var err error
			var timeDate []byte
			timeDate, err = utils.FormatBinaryTime(int(length), rows[pos:])
			row[index] = string(timeDate)
			if err != nil {
				glog.Error(err)
				return nil
			}
			pos += int(length)
			continue
			//todo
		default:
			glog.Errorf("Stmt Unknown FieldType %d %s", field.typ, field.name)
			return nil
		}
	}
	return row
}
