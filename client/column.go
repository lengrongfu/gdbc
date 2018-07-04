package client

import (
	"encoding/binary"
	"gdbc/constant"
	"gdbc/packets"
	"gdbc/utils"
	"github.com/golang/glog"
)

//ColumnDef41Packets 结构
type ColumnDef41Packets struct {
	catalog            string
	schema             string
	table              string
	orgTable           string
	name               string
	orgName            string
	fixedLength        uint8 //length of the following fields (always 0x0c)
	characterSet       uint16
	columnLength       uint32
	typ                uint8
	flags              uint16
	decimals           uint8
	filler             uint16
	defaultValueLength uint64
	defaultValue       string
}

//HandlerColumnDef41 ...
func HandlerColumnDef41(payload []byte) ColumnDef41Packets {
	var index, n int
	column := ColumnDef41Packets{}
	column.catalog, n, _ = utils.LenencStringDecode(payload[index:])
	index += n
	column.schema, n, _ = utils.LenencStringDecode(payload[index:])
	index += n
	column.table, n, _ = utils.LenencStringDecode(payload[index:])
	index += n
	column.orgTable, n, _ = utils.LenencStringDecode(payload[index:])
	index += n
	column.name, n, _ = utils.LenencStringDecode(payload[index:])
	index += n
	column.orgName, n, _ = utils.LenencStringDecode(payload[index:])
	index += n
	column.fixedLength = byte(payload[index])
	index++
	column.characterSet = binary.LittleEndian.Uint16(payload[index:])
	index += 2
	column.columnLength = binary.LittleEndian.Uint32(payload[index:])
	index += 4
	column.typ = byte(payload[index])
	index++
	column.flags = binary.LittleEndian.Uint16(payload[index:])
	index += 2
	column.decimals = byte(payload[index])
	index++
	column.filler = binary.LittleEndian.Uint16(payload[index:])
	index += 2
	if len(payload) > index {
		column.defaultValueLength = binary.LittleEndian.Uint64(payload[index:])
		index += 8
		column.defaultValue = string(payload[index:])
	}
	return column
}

//HandlerColumnDef41s 处理结果集
func HandlerColumnDef41s(c *Conn) ([]ColumnDef41Packets, error) {
	columns := make([]ColumnDef41Packets, 0)
	for {
		//next packets
		if err := c.Payload.ReadPayload(c.netConn); err != nil {
			glog.Errorf("read COM_STMT_EXECUTE column result error:%+v", err)
			return nil, err
		}
		payload := c.Payload.Payload
		capabilities := c.Payload.Capability
		//If the CLIENT_DEPRECATE_EOF client capability flag is set, OK_Packet is sent; else EOF_Packet is sent.
		if capabilities&constant.ClientDeprecateEOF == 0 {
			if eofPacket := packets.EOFPacketHandler(payload, capabilities); eofPacket.IsEOF() {
				break
			}
		} else {
			if okPackt := packets.OkPacketHandler(payload, capabilities); okPackt.IsOk() {
				break
			}
		}
		columns = append(columns, HandlerColumnDef41(payload))
	}
	return columns, nil
}
