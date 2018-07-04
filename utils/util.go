// create_by : lengrongfu
package utils

import (
	"database/sql/driver"
	"encoding/binary"
	"fmt"
	"gdbc/constant"
	"math"
	"strconv"
	"time"
)

/**
大端模式，是指数据的高字节保存在内存的低地址中，而数据的低字节保存在内存的高地址中，这样的存储模式有点儿类似于把数据当作字符串顺序处理：地址由小向大增加，而数据从高位往低位放；这和我们的阅读习惯一致。
小端模式，是指数据的高字节保存在内存的高地址中，而数据的低字节保存在内存的低地址中，这种存储模式将地址的高低和数据位权有效地结合起来，高地址部分权值高，低地址部分权值低。
*/
// LenencIntEncoded 解码可变长度的数据,解码数据
// https://dev.mysql.com/doc/internals/en/integer.html#packet-Protocol::LengthEncodedInteger
func LenencIntDecode(data []byte) (num uint64, n int) {
	// 若 N1 <= 0xfb, 则说明该整数就是N1
	if data[0] <= 0xfb {
		n = 1
		num = uint64(data[0])
		return
	}
	switch data[0] {
	case 0xfc:
		n = 3
		//mysql 协议采用的是小端模式，加法运算可以采用按位或
		num = uint64(data[1]) | uint64(data[2])<<8
		return
	case 0xfd:
		n = 4
		num = uint64(1) | uint64(2)<<8 | uint64(data[3])<<16
		return
	case 0xfe:
		n = 9
		num = uint64(1) | uint64(2)<<8 | uint64(data[3])<<16 | uint64(data[4])<<24 |
			uint64(data[5])<<32 | uint64(data[6])<<40 | uint64(data[7])<<48 |
			uint64(data[8])<<56
		return
	}
	return
}

/**
To convert a number value into a length-encoded integer:

If the value is < 251, it is stored as a 1-byte integer.

If the value is ≥ 251 and < (2^16), it is stored as fc + 2-byte integer.

If the value is ≥ (2^16) and < (2^24), it is stored as fd + 3-byte integer.

If the value is ≥ (2^24) and < (2^64) it is stored as fe + 8-byte integer.
*/
//LenencIntEncode mysql 可变长度数据进行编码
// https://dev.mysql.com/doc/internals/en/integer.html#packet-Protocol::LengthEncodedInteger
func LenencIntEncode(data uint64) (value []byte) {
	if data < 251 {
		value = []byte{uint8(data)}
		return
	} else if data < 0xffff {
		value = []byte{0xfc, uint8(data), uint8(data) >> 8}
		return
	} else if data < 0xffffffff {
		value = []byte{0xfd, uint8(data), uint8(8) >> 8, uint8(data) >> 16}
		return
	} else if data < 0xffffffffffffffff {
		value = []byte{0xfe, uint8(data), uint8(data) >> 8, uint8(data) >> 16, uint8(data) >> 24,
			uint8(data) >> 32, uint8(data) >> 40, uint8(data) >> 48, uint8(data) >> 56}
		return
	}
	return nil
}

// NULL is sent as 0xfb
//LenencStringDecode 可变长的字符串
func LenencStringDecode(data []byte) (value string, n int, isNull bool) {
	var length uint64
	length, n = LenencIntDecode(data)
	if length == 0xfb {
		isNull = true
		return
	}
	isNull = false
	value = string(data[n:(n + int(length))])
	n += len(value)
	return
}

const digits01 = "0123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789"
const digits10 = "0000000000111111111122222222223333333333444444444455555555556666666666777777777788888888889999999999"

//DriverValueHandler value type change
func DriverValueHandler(arg driver.Value, protocol int) (paramType [2]byte, data []byte) {
	if protocol != constant.TextProtocol && protocol != constant.BinaryProtocol {
		return
	}
	switch v := arg.(type) {
	// case int8:
	// 	paramType[0] = constant.MysqlTypeTiny
	// 	return paramType, []byte{byte(v)}
	// case int16:
	// 	paramType[0] = constant.MysqlTypeShort
	// 	data = make([]byte, 2)
	// 	binary.LittleEndian.PutUint16(data, uint16(v))
	// 	return paramType, data
	// case int32:
	// 	paramType[0] = constant.MysqlTypeLong
	// 	data = make([]byte, 4)
	// 	binary.LittleEndian.PutUint32(data, uint32(v))
	// 	return paramType, data
	// case int:
	// 	// /strconv.AppendInt(data, int64(v), 10)
	// 	paramType[0] = constant.MysqlTypeLongLong
	// 	data = make([]byte, 8)
	// 	binary.LittleEndian.PutUint64(data, uint64(v))
	// 	return paramType, data
	// case uint8:
	// 	paramType[0] = constant.MysqlTypeTiny
	// 	paramType[1] = 0x80
	// 	return paramType, []byte{v}
	// case uint16:
	// 	paramType[0] = constant.MysqlTypeShort
	// 	paramType[1] = 0x80
	// 	data = make([]byte, 2)
	// 	binary.LittleEndian.PutUint16(data, v)
	// 	return paramType, data
	// case uint32:
	// 	paramType[0] = constant.MysqlTypeLong
	// 	paramType[1] = 0x80
	// 	data = make([]byte, 4)
	// 	binary.LittleEndian.PutUint32(data, v)
	// 	return paramType, data
	// case uint:
	// 	paramType[0] = constant.MysqlTypeLongLong
	// 	paramType[1] = 0x80
	// 	return paramType, strconv.AppendInt(data, int64(v), 10)
	// case uint64:
	// 	paramType[0] = constant.MysqlTypeLongLong
	// 	paramType[1] = 0x80
	// 	return paramType, strconv.AppendInt(data, int64(v), 10)
	case int64:
		if protocol == constant.TextProtocol {
			data = strconv.AppendInt(data, v, 10)
		} else if protocol == constant.BinaryProtocol {
			paramType[0] = constant.MysqlTypeLongLong
			data = make([]byte, 8)
			binary.LittleEndian.PutUint64(data, uint64(v))
		}
		return paramType, data
	case float64:
		if protocol == constant.TextProtocol {
			data = strconv.AppendFloat(data, v, 'g', -1, 64)
		} else if protocol == constant.BinaryProtocol {
			paramType[0] = constant.MysqlTypeDouble
			binary.LittleEndian.PutUint64(data, math.Float64bits(v))
		}
		return
	case bool:
		if protocol == constant.BinaryProtocol {
			paramType[0] = constant.MysqlTypeTiny
		}
		if v {
			data = append(data, '1')
		} else {
			data = append(data, '0')
		}
		return
	case []byte:
		if protocol == constant.TextProtocol {
			if v == nil {
				data = append(data, "NULL"...)
			} else {
				data = v
			}
		} else if protocol == constant.BinaryProtocol {
			paramType[0] = constant.MysqlTypeString
			data = append(LenencIntEncode(uint64(len(v))), v...)
		}
		return
	case string:
		if protocol == constant.TextProtocol {
			data = []byte(v)
		} else if protocol == constant.BinaryProtocol {
			paramType[0] = constant.MysqlTypeString
			data = append(LenencIntEncode(uint64(len(v))), v...)
		}
		return
	case time.Time:
		if protocol == constant.TextProtocol {
			if v.IsZero() {
				return paramType, []byte("'0000-00-00'")
			} else {
				v := v.In(time.Local)
				v = v.Add(time.Nanosecond * 500) // To round under microsecond
				year := v.Year()
				year100 := year / 100
				year1 := year % 100
				month := v.Month()
				day := v.Day()
				hour := v.Hour()
				minute := v.Minute()
				second := v.Second()
				micro := v.Nanosecond() / 1000

				data = append(data, []byte{
					'\'',
					digits10[year100], digits01[year100],
					digits10[year1], digits01[year1],
					'-',
					digits10[month], digits01[month],
					'-',
					digits10[day], digits01[day],
					' ',
					digits10[hour], digits01[hour],
					':',
					digits10[minute], digits01[minute],
					':',
					digits10[second], digits01[second],
				}...)

				if micro != 0 {
					micro10000 := micro / 10000
					micro100 := micro / 100 % 100
					micro1 := micro % 100
					data = append(data, []byte{
						'.',
						digits10[micro10000], digits01[micro10000],
						digits10[micro100], digits01[micro100],
						digits10[micro1], digits01[micro1],
					}...)
				}
				data = append(data, '\'')
			}
			return
		} else if protocol == constant.BinaryProtocol {
			paramType[0] = constant.MysqlTypeTime
			if v.Year() == 0 && v.Month() == 0 && v.Day() == 0 && v.Hour() == 0 &&
				v.Minute() == 0 && v.Second() == 0 && v.Nanosecond() == 0 {
				data = []byte{0}
				return
			}
			if v.Hour() == 0 && v.Minute() == 0 && v.Second() == 0 && v.Nanosecond() == 0 {
				year := make([]byte, 2)
				binary.LittleEndian.PutUint16(year, uint16(v.Year()))
				data = []byte{0x04}
				data = append(data, year...)
				temp := []byte{byte(v.Month()), byte(v.Day())}
				data = append(data, temp...)
				return
			}
			if v.Nanosecond() == 0 {
				year := make([]byte, 2)
				binary.LittleEndian.PutUint16(year, uint16(v.Year()))
				data = []byte{0x04}
				data = append(data, year...)
				temp := []byte{byte(v.Month()), byte(v.Day()), byte(v.Hour()), byte(v.Minute()), byte(v.Second())}
				data = append(data, temp...)
				return
			}
			year := make([]byte, 2)
			binary.LittleEndian.PutUint16(year, uint16(v.Year()))
			data = []byte{0x04}
			data = append(data, year...)
			temp := []byte{byte(v.Month()), byte(v.Day()), byte(v.Hour()), byte(v.Minute()), byte(v.Second())}
			data = append(data, temp...)
			ns := make([]byte, 4)
			binary.LittleEndian.PutUint32(ns, uint32(v.Nanosecond()))
			data = append(data, ns...)
			return
		}
	default:
		return paramType, []byte("NULL")
	}
	return paramType, nil
}

//FormatBinaryDate format
func FormatBinaryDate(n int, data []byte) ([]byte, error) {
	switch n {
	case 0:
		return []byte("0000-00-00"), nil
	case 4:
		return []byte(fmt.Sprintf("%04d-%02d-%02d",
			binary.LittleEndian.Uint16(data[:2]),
			data[2],
			data[3])), nil
	default:
		return nil, fmt.Errorf("invalid date packet length %d", n)
	}
}

// FormatBinaryDateTime ...
func FormatBinaryDateTime(n int, data []byte) (time.Time, error) {
	data = data[1:]
	switch n {
	case 0:
		return time.Parse("2006-01-02 00:00:00", "0000-00-00 00:00:00")
	case 4:
		tstr := fmt.Sprintf("%04d-%02d-%02d 00:00:00",
			binary.LittleEndian.Uint16(data[:2]),
			data[2],
			data[3])
		return time.Parse("2006-01-02 00:00:00", tstr)
	case 7:
		tstr := fmt.Sprintf(
			"%04d-%02d-%02d %02d:%02d:%02d",
			binary.LittleEndian.Uint16(data[:2]),
			data[2],
			data[3],
			data[4],
			data[5],
			data[6])
		return time.Parse("2006-01-02 00:00:00", tstr)
	case 11:
		tstr := fmt.Sprintf(
			"%04d-%02d-%02d %02d:%02d:%02d.%06d",
			binary.LittleEndian.Uint16(data[:2]),
			data[2],
			data[3],
			data[4],
			data[5],
			data[6],
			binary.LittleEndian.Uint32(data[7:11]))
		return time.Parse("2006-01-02 00:00:00.000", tstr)
	default:
		return time.Now(), fmt.Errorf("invalid datetime packet length %d", n)
	}
}

//FormatBinaryTime ...
func FormatBinaryTime(n int, data []byte) ([]byte, error) {
	if n == 0 {
		return []byte("0000-00-00"), nil
	}

	var sign byte
	if data[0] == 1 {
		sign = byte('-')
	}

	switch n {
	case 8:
		return []byte(fmt.Sprintf(
			"%c%02d:%02d:%02d",
			sign,
			uint16(data[1])*24+uint16(data[5]),
			data[6],
			data[7],
		)), nil
	case 12:
		return []byte(fmt.Sprintf(
			"%c%02d:%02d:%02d.%06d",
			sign,
			uint16(data[1])*24+uint16(data[5]),
			data[6],
			data[7],
			binary.LittleEndian.Uint32(data[8:12]),
		)), nil
	default:
		return nil, fmt.Errorf("invalid time packet length %d", n)
	}
}
