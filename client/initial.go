package client

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"gdbc/constant"
	"gdbc/packets"
)

// https://dev.mysql.com/doc/internals/en/initial-handshake.html
// readInitialHandShake : handshake first step,read mysql server return handshake packets
func (c *Conn) readInitialHandShake() error {
	if err := c.Payload.ReadPayload(c.netConn); err != nil {
		return err
	}
	payload := c.Payload.Payload

	if errPackt := packets.ErrPacketHandler(payload, c.Payload.Capability); errPackt.IsErr() {
		return fmt.Errorf("read initialhandshake error:%s", errPackt.String())
	}
	if payload[0] < constant.MinProtocolVersion {
		invalidVersion := fmt.Sprintf("invalid protocol version %d ,version is must >= 10.", payload[0])
		return errors.New(invalidVersion)
	}
	// skip server version
	// server version end with 0x00
	// 1 : protocolVersion length; bytes.IndexByte(Payload,0x00) server version length;1 0x00 byte
	// index := 1 + bytes.IndexByte(payload[1:], 0x00) + 1
	index := 1
	c.serverVersion = string(payload[1:bytes.IndexByte(payload[1:], 0x00)])
	index += bytes.IndexByte(payload[1:], 0x00) + 1
	//4 connection id
	c.connectionID = binary.LittleEndian.Uint32(payload[index : index+4])
	index += 4
	//string[8] auth-plugin-data-part-1
	authData := []byte{}
	authData = append(authData, payload[index:index+8]...)
	index += 8
	//1              [00] filler ;skip
	index ++
	//2              capability flags (lower 2 bytes)
	c.Payload.Capability = uint32(binary.LittleEndian.Uint16(payload[index : index+2]))
	index += 2
	//if more data in the packet
	if len(payload) > index {
		//1              character set
		c.characterSet = payload[index]
		index ++
		//2              status flags
		c.statusFlags = binary.LittleEndian.Uint16(payload[index:])
		index += 2
		//2              capability flags (upper 2 bytes)
		c.Payload.Capability = uint32(binary.LittleEndian.Uint16(payload[index:index+2]))<<16 | c.Payload.Capability
		index += 2
		//1              length of auth-plugin-data
		/**
		var authDateLength uint8
		if c.capability & constant.ClientPluginAuth != 0{
			authDateLength = uint8(Payload[index:index+1][0])
		}else{
			authDateLength = uint8(0x00)
		}
		*/
		index ++
		//string[10]     reserved (all [00])
		index += 10
		authData = append(authData, payload[index:index+12]...)
		index += 13
		c.authData = authData
		// auth plugin name
		var plugin string
		if end := bytes.IndexByte(payload[index:], 0x00); end != -1 {
			plugin = string(payload[index : index+end])
		} else {
			plugin = string(payload[index:])
		}
		c.authPluginName = []byte(plugin)

	}
	return nil
}
