// create_by : lengrongfu
package packets

import (
	"encoding/binary"
	"fmt"
	"gdbc/constant"
)

// ErrPacket packets
type ErrPacket struct {
	err            bool
	errorCode      uint16 //int<2>	error_code	error-code
	sqlStateMarker uint8  //marker of the SQL State
	sqlState       []byte // SQL State 5
	errorMessage   string // human readable error message string<EOF> eof 数据为剩余的数据
}

//ErrPacketHandler err packet 处理
func ErrPacketHandler(payload []byte, capabilities uint32) ErrPacket {
	errP := ErrPacket{
		err: false,
	}
	index := 0
	if payload[index] != constant.ErrHeader {
		return errP
	}
	errP.err = true
	index++
	if capabilities&constant.ClientProtocol41 > 0 {
		errP.errorCode = binary.LittleEndian.Uint16(payload[index:])
		index += 2
		errP.sqlStateMarker = uint8(payload[index])
		index++
		errP.sqlState = payload[index : index+5]
		index += 5
	}
	errP.errorMessage = string(payload[index:])
	return errP
}

//IsErr is return err
func (e ErrPacket) IsErr() bool {
	return e.err
}

//String is implement String interface
func (p ErrPacket) String() string {
	return fmt.Sprintf("error_code is %v,SQL State is %v ,error message is %s ", p.errorCode, string(p.sqlState), p.errorMessage)
}
