package client

import (
	"crypto/tls"
	"gdbc/constant"
	"gdbc/gdbcerrors"
	"github.com/golang/glog"
	"strconv"
	"strings"
)

//Debug 开启调试
var Debug bool

// mysql client config
type config struct {
	User        string            //用户名
	Password    string            //密码
	Host        string            //ip or domain name
	Port        int               // port
	DB          string            // Client DB
	TLSConfig   *tls.Config       //是否使用ssl连接
	charset     string            //连接字符编码
	params      map[string]string //其它参数
	DEBUG       bool              //是否开启调试，未使用
	Compression bool              //是否压缩
}

//ResolveDSNName check and resolve dsnname
func ResolveDSNName(dsnName string) (*config, error) {
	cnf := new(config)
	cnf.DEBUG = Debug
	dsn := []byte(dsnName)
	usAndPwd := strings.LastIndex(dsnName, "@")
	userAndAddrs := []string{string(dsn[:usAndPwd]), string(dsn[usAndPwd+1:])}
	if len(userAndAddrs) != 2 {
		glog.Errorln("dsnName is invalid:", dsnName)
		return nil, gdbcerrors.ErrDsn
	}
	//user pwd
	userAndPassword := strings.Split(userAndAddrs[0], ":")
	if len(userAndPassword) < 1 {
		glog.Errorln("dsnName is invalid:", dsnName)
		return nil, gdbcerrors.ErrDsn
	} else if len(userAndPassword) == 1 {
		cnf.User = userAndPassword[0]
	} else {
		cnf.User = userAndPassword[0]
		cnf.Password = userAndPassword[1]
	}
	//addr db
	addrsAndDB := strings.Split(userAndAddrs[1], "/")
	if len(addrsAndDB) != 2 {
		glog.Errorln("dsnName is invalid:", dsnName)
		return nil, gdbcerrors.ErrDsn
	}
	hostAndPort := strings.Split(addrsAndDB[0], ":")
	if len(hostAndPort) != 2 {
		glog.Errorln("dsnName is invalid:", dsnName)
		return nil, gdbcerrors.ErrDsn
	}

	cnf.Host = hostAndPort[0]
	cnf.Port, _ = strconv.Atoi(hostAndPort[1])

	//db param
	dbAndParams := strings.Split(addrsAndDB[1], "?")
	if len(dbAndParams) == 1 {
		cnf.DB = dbAndParams[0]
	} else {
		param := parseParam(dbAndParams[1])
		if param != nil {
			cnf.params = param
		}
	}

	cnf.charset = constant.DefaultCharset

	return cnf, nil
}

// parseParam parse dsn in param
func parseParam(params string) map[string]string {
	if params == "" {
		return nil
	}
	keyValues := strings.Split(params, "&")
	param := make(map[string]string, len(keyValues))
	for _, keyValue := range keyValues {
		keyAndValue := strings.Split(keyValue, "=")
		param[keyAndValue[0]] = keyAndValue[1]
	}
	return param
}
