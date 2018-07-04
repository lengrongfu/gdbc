package client

import (
	"gdbc/constant"
	"github.com/golang/glog"
)

//GdbcTx 是关于事务的操作
type GdbcTx struct {
	c *Conn
}

//Commit 提交事务
func (t GdbcTx) Commit() error {
	if err := t.c.WriteCommandStr(constant.ComQuery, "COMMIT"); err != nil {
		glog.Error(err)
		return err
	}
	return nil
}

// Rollback 事务回滚
func (t GdbcTx) Rollback() error {
	if err := t.c.WriteCommandStr(constant.ComQuery, "ROLLBACK"); err != nil {
		glog.Error(err)
		return err
	}
	return nil
}
