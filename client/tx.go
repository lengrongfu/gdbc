package client

import (
	"errors"
	"fmt"
	"gdbc/constant"
	"gdbc/packets"
	"github.com/golang/glog"
)

//GdbcTx 是关于事务的操作
type GdbcTx struct {
	c *Conn
}

//Commit 提交事务
func (t GdbcTx) Commit() error {
	if t.c == nil {
		return errors.New("transaction client is close")
	}
	if err := t.c.WriteCommandStr(constant.ComQuery, "COMMIT"); err != nil {
		glog.Error(err)
		return err
	}
	if readErr := t.c.Payload.ReadPayload(t.c.netConn); readErr != nil {
		glog.Errorf("COMMIT TRANSACTION,Reader result error:%+v", readErr)
		return fmt.Errorf("COMMIT TRANSACTION,Reader result error:%s", readErr.Error())
	}
	paylod := t.c.Payload.Payload
	capability := t.c.Payload.Capability
	if errPack := packets.ErrPacketHandler(paylod, capability); errPack.IsErr() {
		glog.Errorf("WHEN COMMIT TRANSACTION ,MYSQL SERVER RETURN ERROR:%s", errPack.String())
		return errors.New(errPack.String())
	}
	if okPacket := packets.OkPacketHandler(paylod, capability); okPacket.IsOk() {
		return nil
	}
	return errors.New("Commit Transaction unknown mistake")
}

// Rollback 事务回滚
func (t GdbcTx) Rollback() error {
	if t.c == nil {
		return nil
	}
	if err := t.c.WriteCommandStr(constant.ComQuery, "ROLLBACK"); err != nil {
		glog.Error(err)
		return err
	}
	if readErr := t.c.Payload.ReadPayload(t.c.netConn); readErr != nil {
		glog.Errorf("ROLLBACK TRANSACTION,Reader result error:%+v", readErr)
		return fmt.Errorf("ROLLBACK TRANSACTION,Reader result error:%s", readErr.Error())
	}
	paylod := t.c.Payload.Payload
	capability := t.c.Payload.Capability
	if errPack := packets.ErrPacketHandler(paylod, capability); errPack.IsErr() {
		glog.Errorf("WHEN ROLLBACK TRANSACTION ,MYSQL SERVER RETURN ERROR:%s", errPack.String())
		return errors.New(errPack.String())
	}
	if okPacket := packets.OkPacketHandler(paylod, capability); okPacket.IsOk() {
		return nil
	}
	return errors.New("Rollback Transaction unknown mistake")
}
