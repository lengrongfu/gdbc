package client

import (
	"crypto/tls"
	"gdbc/constant"
	"github.com/golang/glog"
)

// sslExchange:https://dev.mysql.com/doc/internals/en/ssl-handshake.html
//  https://dev.mysql.com/doc/internals/en/ssl.html
func (c *Conn) sslExchange() error {
	capability := constant.ClientSsl
	// 4 PayloadLength [3]byte SequenceID    [1]byte
	// 4 capability flags
	// 4 max-packet size
	// 1 character set
	// 23 reserved (all [0])
	data := make([]byte, 4+(4+4+1+23))
	data[4] = byte(capability)
	data[5] = byte(capability >> 8)
	data[6] = byte(capability >> 16)
	data[7] = byte(capability >> 24)
	//use default character ,33 is utf8
	data[12] = byte(constant.DefaultCharsetByte)
	if err := c.WritePacket(data); err != nil {
		glog.Errorf("writepacket error:%#v\n", err)
		return err
	}
	tlsConn := tls.Client(c.netConn, c.cnf.TLSConfig)
	if err := tlsConn.Handshake(); err != nil {
		glog.Errorf("tls connection is error:%#v\n", err)
		return err
	}
	c.netConn = tlsConn
	return nil
}
