// create_by : lengrongfu
// This package implements databases/sql/driver interface,so we can user gdbc with databases/sql
package gdbc

import (
	"database/sql"
	"database/sql/driver"
	"gdbc/client"
	"github.com/golang/glog"
)

//mysqlDrive
type mysqlDrive struct {
}

// dsnName : {user}:{password}@{addr}/{dbName}?paramsKey1={paramsValue1}&paramsKey2={paramsValues2}
func (mysql mysqlDrive) Open(dsnName string) (driver.Conn, error) {
	cnf, err := client.ResolveDSNName(dsnName)
	if err != nil {
		glog.Error(err.Error())
		return nil, err
	}
	if cnf.DEBUG {
		glog.Infof("dsn resolve success:%+v", cnf)
	}
	conn, err := client.Connect(*cnf)
	if err != nil {
		return nil, err
	}
	if cnf.DEBUG {
		glog.Info("gdbc connection mysql servier success .....")
	}
	return conn, nil
}

// init mysql drive register
func init() {
	sql.Register("mysql", &mysqlDrive{})
}
