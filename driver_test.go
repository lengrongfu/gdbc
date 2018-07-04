package gdbc

import (
	"testing"
)

func Test_Open(t *testing.T) {
	mysqlD := new(mysqlDrive)
	dsnName := "root:123456@127.0.0.1:3306/test"
	c, err := mysqlD.Open(dsnName)
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	t.Log(c)
}
