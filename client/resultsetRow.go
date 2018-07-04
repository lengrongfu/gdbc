package client

import (
	"gdbc/constant"
	"gdbc/packets"
	"github.com/golang/glog"
)

//BinaryResultsetRow is Binary Protocol Resultset
type BinaryResultsetRow struct {
	header uint8
	//http://blog.jobbole.com/94166/ NullBitmap 讲解
	NULLBitmap []byte
	values     []byte
}

//HandlerResultSetRow is parse pyload to BinaryResultsetRow
func HandlerResultSetRow(payload []byte, columnCount uint64) BinaryResultsetRow {
	result := BinaryResultsetRow{}
	result.header = payload[0]
	offset := 2 //For the Binary Protocol Resultset Row the num-fields and the field-pos need to add a offset of 2. For COM_STMT_EXECUTE this offset is 0.
	NULLBitmapBytes := (columnCount + 7 + uint64(offset)) >> 3
	if NULLBitmapBytes > 0 {
		//这是生成规则，客户端直读取就行
		//result.NULLBitmap = make([]byte,NULLBitmapBytes)
		//for fieldPos := 0; fieldPos < int(columnCount); fieldPos++ {
		//	NULLBitmapByte := ((fieldPos + offset) / 8)
		//	NULLBitmapBit  := ((fieldPos + offset) % 8)
		//	result.NULLBitmap[NULLBitmapByte] |= 1 << uint(NULLBitmapBit)
		//}
		result.NULLBitmap = payload[1 : NULLBitmapBytes+1]
	}
	//values is a string<lenenc_str>
	result.values = payload[len(result.NULLBitmap)+1:]
	return result
}

// HandlerResultSetRowS ....
func HandlerResultSetRowS(c *Conn, columnCount uint64) ([]BinaryResultsetRow, error) {
	results := make([]BinaryResultsetRow, 0)
	for {
		if err := c.Payload.ReadPayload(c.netConn); err != nil {
			glog.Errorf("read COM_STMT_EXECUTE result error:%+v", err)
			return nil, err
		}
		payload := c.Payload.Payload
		capabilities := c.Payload.Capability
		//If the CLIENT_DEPRECATE_EOF client capability flag is set, OK_Packet is sent; else EOF_Packet is sent.
		if c.Payload.Capability&constant.ClientDeprecateEOF == 0 {
			if eofPacket := packets.EOFPacketHandler(payload, capabilities); eofPacket.IsEOF() {
				break
			}
		} else {
			if okPackt := packets.OkPacketHandler(payload, capabilities); okPackt.IsOk() {
				break
			}
		}
		result := HandlerResultSetRow(payload, columnCount)
		results = append(results, result)
	}
	return results, nil
}

// NULL is sent as 0xfb
// everything else is converted into a string and is sent as Protocol::LengthEncodedString.
type TextResultsetRow struct {
	isNull bool
	values string
}
