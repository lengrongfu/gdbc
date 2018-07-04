package client

import (
	"crypto/sha1"
	"errors"
	"gdbc/constant"
	"gdbc/gdbcerrors"
	"gdbc/packets"
	"github.com/golang/glog"
)

//writeAuthenticationData :https://dev.mysql.com/doc/internals/en/connection-phase-packets.html#packet-Protocol::HandshakeResponse

func (c *Conn) writeAuthenticationData() error {
	capability := constant.ClientProtocol41 | constant.ClientSecureConnection |
		constant.ClientLongPassword | constant.ClientTransactions | constant.ClientLongFlag
	//c.capability 就是mysql服务器支持的功能的标志位，按位与上我们需要的功能，如果服务器不支持则不能使用此功能
	capability &= c.Payload.Capability
	//packet length
	//capbility 4
	//max-packet size 4
	//charset 1
	//reserved all[0] 23
	length := 4 + 4 + 1 + 23
	//username
	length += len(c.cnf.User) + 1
	//default support secure connection
	authResp, err := c.calcAuth()
	if err != nil {
		glog.Infof("auth plugin calc error:%v\n", err)
		authResp = defaultAuth(c.authData, []byte(c.cnf.Password))
	}
	length += len(authResp) + 1
	//db
	if len(c.cnf.DB) > 0 {
		capability |= constant.ClientConnectWithDb
		length += len(c.cnf.DB) + 1
	}
	//压缩
	if c.cnf.Compression {
		capability |= constant.ClientCompress
	}
	// mysql_native_password + null-terminated
	length += 21 + 1
	c.Payload.Capability = capability
	data := make([]byte, length+4)

	//capability [32 bit]
	data[4] = byte(capability)
	data[5] = byte(capability >> 8)
	data[6] = byte(capability >> 16)
	data[7] = byte(capability >> 24)
	data[12] = byte(constant.DefaultCharsetByte)
	//Filler [23 bytes] (all 0x00)
	pos := 13 + 23
	//User [null terminated string]
	if len(c.cnf.User) > 0 {
		pos += copy(data[pos:], c.cnf.User)
	}
	data[pos] = 0x00
	pos++

	// auth [length encoded integer]
	data[pos] = byte(len(authResp))
	pos += 1 + copy(data[pos+1:], authResp)

	// db [null terminated string]
	if len(c.cnf.DB) > 0 {
		pos += copy(data[pos:], c.cnf.DB)
		data[pos] = 0x00
		pos++
	}

	// Assume native client during response
	pos += copy(data[pos:], c.authPluginName)
	data[pos] = 0x00

	return c.WritePacket(data)
}

//默认认证方式
func defaultAuth(authData []byte, password []byte) []byte {
	return scrambleNativePassword(authData, password)
}

//计算auth_plugin 和 password 的数据
func (c Conn) calcAuth() ([]byte, error) {
	authPlugin := string(c.authPluginName)
	auth := make([]byte, 0)
	switch authPlugin {
	case constant.MysqlNativePassword:
		if c.Payload.Capability&constant.ClientSecureConnection != 0 {
			return scrambleNativePassword(c.authData, []byte(c.cnf.Password)), nil
		}
	case constant.MysqlOldPassword:
		if 1 > 0 {
			glog.Errorf(" %s auth plugin no secure!\n", authPlugin)
			return nil, gdbcerrors.ErrNoSecureAuthPlugin
		}
		//次插件已经不安全被破解，所以不再支持
		clientSide := make([]byte, 0)
		if c.Payload.Capability&constant.ClientSecureConnection != 0 {
			copy(c.authData[:8], clientSide)
		} else {
			clientSide = nil
		}
		return clientSide, nil
	case constant.MysqlClearPassword:
		return []byte(c.cnf.Password), nil
	case constant.AuthenticationWindowsClient:
		if 1 > 0 {
			return nil, errors.New("sorry this auth plugin is not support")
		}
	case constant.Sha256Password:
		if 1 > 0 {
			return nil, errors.New("sorry this auth plugin is not support")
		}
		if c.cnf.Password == "" {
			return []byte{}, nil
		}
	default:
		glog.Errorf("no support this auth plugin:%s\n", authPlugin)
		return nil, gdbcerrors.ErrAuthPluginNoSupport
	}

	return auth, nil
}

// https://dev.mysql.com/doc/internals/en/secure-password-authentication.html
//SHA1( password ) XOR SHA1( "20-bytes random data from server" <concat> SHA1( SHA1( password ) ) ) ;<concat> 是拼接字符的标示
//scrambleNativePassword is mysql_native_password scramble
func scrambleNativePassword(authData []byte, password []byte) []byte {
	//SHA1( password )
	crypt := sha1.New()
	crypt.Write(password)
	sha1Pwd := crypt.Sum(nil)
	//SHA1( SHA1( password ) )
	sha1Sha1 := sha1.New()
	sha1Sha1.Write(sha1Pwd)
	sha1Sha1Pwd := sha1Sha1.Sum(nil)
	//
	sha1Concat := sha1.New()
	sha1Concat.Write(authData)
	sha1Concat.Write(sha1Sha1Pwd)
	sha1ConcatPwd := sha1Concat.Sum(nil)
	//xop
	for i := range sha1ConcatPwd {
		sha1ConcatPwd[i] = sha1ConcatPwd[i] ^ sha1Pwd[i]
	}
	return sha1ConcatPwd
}

// 处理返回结果
func (c *Conn) handleAuthResult() error {
	if err := c.Payload.ReadPayload(c.netConn); err != nil {
		return err
	}
	switch c.Payload.Payload[0] {
	case constant.OkHeader:
		return nil
	case 0x01:
		//The packets which server sends in step 4 are the Extra Authentication Data packet prefixed with 0x01
		glog.Error("在这个认证方式中还需要额外的数据。")
		return errors.New("need Extra Authentication Data packet in this auth")
	case constant.ErrHeader:
		errP := packets.ErrPacketHandler(c.Payload.Payload, c.Payload.Capability)
		return errors.New(errP.String())
	default:
		return errors.New("response packet header invalid ")
	}
}
