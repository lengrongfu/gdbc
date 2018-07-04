// create_by : lengrongfu
package packets

import (
	"encoding/binary"
	"fmt"
	"gdbc/constant"
)

//EOFPacket eof packets
type EOFPacket struct {
	eof         bool
	warnings    uint16
	statusFlags uint16 //Status Flags
}

//EOFPacketHandler 判断是否是错误包
func EOFPacketHandler(payload []byte, capabilities uint32) EOFPacket {
	eofP := EOFPacket{
		eof: false,
	}
	if payload[0] != constant.EofHeader {
		return eofP
	}
	eofP.eof = true
	if capabilities&constant.ClientProtocol41 > 0 {
		//uint16(payload[1]) | uint16(payload[2]) << 8
		eofP.warnings = binary.LittleEndian.Uint16(payload[1:])
		eofP.statusFlags = binary.LittleEndian.Uint16(payload[3:])
	}
	return eofP
}

//IsEOF 是否是eof 包
func (eof EOFPacket) IsEOF() bool {
	return eof.eof
}

func (eof EOFPacket) String() string {
	return fmt.Sprintf("EOF packets ,warnings lines is %d,Status Flags is %d", eof.warnings, eof.statusFlags)
}
