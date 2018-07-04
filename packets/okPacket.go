// create_by : lengrongfu
package packets

import (
	"encoding/binary"
	"gdbc/constant"
	"gdbc/utils"
)

//OKPacket is service send to client ,5.7.5 开始EOF 包不在使用，用OK包代替EOF
// https://dev.mysql.com/doc/internals/en/packet-OK_Packet.html
type OKPacket struct {
	OK                  bool
	Header              byte   //[00] or [fe] the OK packet header
	AffectedRows        uint64 //affected rows
	LastInsertId        uint64 //last insert-id
	StatusFlags         uint16 //Status Flags if capabilities & CLIENT_PROTOCOL_41 > 0
	Warnings            uint16 //number of warnings if capabilities & CLIENT_PROTOCOL_41 > 0
	Info                string //human readable status information if capabilities & CLIENT_SESSION_TRACK
	SessionStateChanges string // session state info

}

//OkPacketHandler method reader send command response
func OkPacketHandler(payload []byte, capabilities uint32) OKPacket {
	index := 0
	okp := OKPacket{
		OK: false,
	}
	okp.Header = payload[index]
	if okp.Header != constant.OkHeader {
		return okp
	}
	okp.OK = true
	if okp.Header == constant.EofHeader && len(payload[1:]) < 8 {
		return okp
	}
	index++
	var n int
	okp.AffectedRows, n = utils.LenencIntDecode(payload[index:])
	index += n
	okp.LastInsertId, n = utils.LenencIntDecode(payload[index:])
	index += n

	if capabilities&constant.ClientProtocol41 > 0 {
		okp.StatusFlags = binary.LittleEndian.Uint16(payload[index:])
		index += 2
		okp.Warnings = binary.LittleEndian.Uint16(payload[index:])
		index += 2
	} else if capabilities&constant.ClientTransactions > 0 {
		okp.StatusFlags = binary.LittleEndian.Uint16(payload[index:])
		index += 2
	}
	if capabilities&constant.ClientSessionTrack > 0 {
		info, n, _ := utils.LenencStringDecode(payload[index:])
		okp.Info = string(info)
		index += n
		if okp.StatusFlags&constant.ServerSessionStateChanged > 0 {
			sessionStateChanges, n, _ := utils.LenencStringDecode(payload[index:])
			okp.SessionStateChanges = string(sessionStateChanges)
			index += n
		}
	} else {
		//string<EOF> 最后的字段为当前长度到最后的字段
		okp.Info = string(payload[index:])
	}
	return okp
}

func (o OKPacket) IsOk() bool {
	return o.OK
}
