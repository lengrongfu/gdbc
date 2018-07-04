package client

import (
	"errors"
	"gdbc/constant"
	"gdbc/packets"
	"github.com/golang/glog"
)

// Quit tells the server that the client wants to close the connection
func (c *Conn) Quit() error {
	defer c.Close()
	com := []byte{constant.ComQuit}
	if err := c.WriteCommand(com); err != nil {
		glog.Errorf("send Quit command error:%v", err)
		return err
	}
	if err := c.Payload.ReadPayload(c.netConn); err != nil {
		return err
	}
	okPack := packets.OkPacketHandler(c.Payload.Payload, c.Payload.Capability)
	if !okPack.OK {
		return errors.New(packets.ErrPacketHandler(c.Payload.Payload, c.Payload.Capability).String())
	}
	return nil
}

// Ping check if the server is alive
// func (c *Conn) Pingc() error {
// 	com := []byte{constant.ComPing}
// 	if err := c.WriteCommand(com); err != nil {
// 		glog.Errorf("send Ping command error:%v", err)
// 		return err
// 	}
// 	if err := c.Payload.ReadPayload(c.netConn); err != nil {
// 		return err
// 	}
// 	okPack := packets.OkPacketHandler(c.Payload.Payload, c.Payload.Capability)
// 	if !okPack.OK {
// 		return errors.New(packets.ErrPacketHandler(c.Payload.Payload, c.Payload.Capability).String())
// 	}
// 	return nil
// }
