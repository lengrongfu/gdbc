// create_by : lengrongfu
//这个包是myslq协议服务端和客户端交互时传输的结构
package packets

import (
	"bytes"
	"database/sql/driver"
	"errors"
	"fmt"
	"gdbc/constant"
	"gdbc/gdbcerrors"
	"github.com/golang/glog"
	"io"
)

// PayloadPackets is payload packet
type PayloadPackets struct {
	PayloadLength [3]byte //payload 长度
	SequenceID    uint8   //交互id，最大255，每次传输最大16M，所以总的最多能传输255*16
	Payload       []byte  //负载数据
	Capability    uint32
}

// PayloadHeader read PayloadLength and SequenceID，Deprecated function
// 读取服务端返回的数据
func (p PayloadPackets) PayloadHeader() error {
	length := int(uint32(p.PayloadLength[0]) | uint32(p.PayloadLength[1])<<8 | uint32(p.PayloadLength[2])<<16)

	if length > constant.MaxPayloadLength {
		glog.Errorln("more than mysql server max packets length.")
		return gdbcerrors.ErrPackentLength
	}

	return nil
}

// ReadPayload method read Payload
func (p *PayloadPackets) ReadPayload(r io.Reader) error {
	var buf bytes.Buffer
	if err := p.ReadPayloadTo(&buf, r); err != nil {
		return err
	}
	p.Payload = buf.Bytes()
	return nil
}

//ReadPayloadTo if length more than 16M,Recursive read
func (p *PayloadPackets) ReadPayloadTo(w io.Writer, r io.Reader) error {
	header := []byte{0, 0, 0, 0}
	n, err := io.ReadFull(r, header)
	if err != nil || n != len(header) {
		glog.Errorf("header read error:%v\n", err)
		return driver.ErrBadConn
	}
	p.PayloadLength = [3]byte{header[0], header[1], header[2]}
	sequence := uint8(header[3])
	if sequence != p.SequenceID {
		invalidSequence := fmt.Sprintf("invalid sequence %d != %d", sequence, p.SequenceID)
		return errors.New(invalidSequence)
	}

	p.SequenceID++

	length := int64(uint32(p.PayloadLength[0]) | uint32(p.PayloadLength[1])<<8 | uint32(p.PayloadLength[2])<<16)
	warnings, err := io.CopyN(w, r, length)
	if err != nil {
		return driver.ErrBadConn
	} else if warnings != length {
		return driver.ErrBadConn
	} else if length < constant.MaxPayloadLength {
		return nil
	} else {
		if err := p.ReadPayloadTo(w, r); err != nil {
			return err
		}
	}
	return nil
}

// RestSequenceID when every send command rest sequenceID 0
func (p *PayloadPackets) RestSequenceID() {
	p.SequenceID = constant.InitSequenceID
}

/**
  校验mysql server 返回的包是不是ok的
   https://dev.mysql.com/doc/internals/en/generic-response-packets.html
   Deprecated function
*/

func (p PayloadPackets) responseStatus() (status int, errCode, errMsg string) {
	if p.Payload == nil || len(p.Payload) == 0 {
		return constant.ERR, "0", "mysql server no return effective data!"
	}
	if p.Payload[0] == constant.OkHeader && len(p.Payload) > constant.OkPacketMinLength {
		return constant.OK, "", ""
	}
	if p.Payload[0] == constant.EofHeader && len(p.Payload) < constant.EofPacketMaxLength {
		return constant.EOF, "0", "mysql server return warnings!"
	}
	if p.Payload[0] == constant.ErrHeader {
		errCode = string(p.Payload[1:3])
		//string<EOF> 是数据包的最后位置，等于总长度减去当前位置
		errMsg = string(p.Payload[9:])
		return constant.ERR, errCode, errMsg
	}
	return constant.ERR, "0", "mysql server packets not match OK,ERR,EOF."
}
