// create_by : lengrongfu
package client

import (
	"context"
	"database/sql/driver"
	"errors"
	"gdbc/constant"
	"gdbc/packets"
	"github.com/golang/glog"
	"net"
	"strconv"
	"time"
)

// Conn phase:https://dev.mysql.com/doc/internals/en/images/graphviz-db6c3eaf9f35f362259756b257b670e75174c29b.png
// Conn is mysql client
type Conn struct {
	netConn        net.Conn
	Payload        *packets.PayloadPackets
	cnf            config
	connectionID   uint32
	authData       []byte
	authPluginName []byte
}

// Connect Constructor
func Connect(cnf config) (*Conn, error) {
	conn := new(Conn)
	conn.cnf = cnf
	conn.Payload = new(packets.PayloadPackets)
	var err error
	addr := cnf.Host + ":" + strconv.Itoa(cnf.Port)
	c, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		return nil, err
	}
	cc, _ := c.(*net.TCPConn)
	cc.SetKeepAlive(true)
	conn.netConn = cc
	if err = conn.handShake(cnf); err != nil {
		return nil, err
	}
	return conn, nil
}

// client handShake : https://dev.mysql.com/doc/internals/en/connection-phase.html
/**
				|   Inital Handshake Packet |
							   |      \
							   |       \
							   |     |SSL exchange|
							   |        /
							   v       /
						   |client response|
						   /          |\
 						  /			  |	 \
						 / 		      |    \
		|Authntication method 		  |      \
			 switch|			      |        \
				/\					  |          \
			   /  \	  			  	  |            \
			  /		\				  |              \
			 /		 \				  |                \
		 	/ 		  \				  |                 /
		   /			-------|	  v                /
      |Disconnect|		 	|Authntication exchange   /
								continuation|        /
									  |   \         /
									  |    \       /
									  |     \     /
									  |      \   /
									  v       \ /
									|OK|     |ERR|
*/
func (c *Conn) handShake(cnf config) error {
	// c.initialHandShake
	if err := c.readInitialHandShake(); err != nil {
		c.Close()
		return err
	}
	//ssl exchange
	if c.cnf.TLSConfig != nil {
		err := c.sslExchange()
		if err != nil {
			c.Close()
			return err
		}
	}
	//client response
	if err := c.writeAuthenticationData(); err != nil {
		c.Close()
		return err
	}
	//server responds
	if err := c.handleAuthResult(); err != nil {
		c.Close()
		return err
	}
	return nil
}

//data already has 4 bytes header
//WritePacket client response service
func (c *Conn) WritePacket(data []byte) error {
	length := len(data) - 4

	for length >= constant.MaxPayloadLength {
		data[0] = 0xff
		data[1] = 0xff
		data[2] = 0xff

		data[3] = c.Payload.SequenceID

		if n, err := c.netConn.Write(data[:4+constant.MaxPayloadLength]); err != nil {
			return driver.ErrBadConn
		} else if n != (4 + constant.MaxPayloadLength) {
			return driver.ErrBadConn
		} else {
			c.Payload.SequenceID++
			length -= constant.MaxPayloadLength
			data = data[constant.MaxPayloadLength:]
		}
	}

	data[0] = byte(length)
	data[1] = byte(length >> 8)
	data[2] = byte(length >> 16)
	data[3] = c.Payload.SequenceID

	if n, err := c.netConn.Write(data); err != nil {
		return driver.ErrBadConn
	} else if n != len(data) {
		return driver.ErrBadConn
	} else {
		c.Payload.SequenceID++
		return nil
	}
}

// WriteCommand send Text command
func (c *Conn) WriteCommand(command []byte) error {
	c.Payload.RestSequenceID()
	comd := []byte{0, 0, 0, 0}
	comd = append(comd, command...)
	if err := c.WritePacket(comd); err != nil {
		glog.Errorf("write command failed,%v", err)
		return err
	}
	return nil
}

//WriteCommandStr send string command
func (c *Conn) WriteCommandStr(command byte, arg string) error {
	c.Payload.RestSequenceID()
	//string<EOF> 编码，最后一个字符串为0x00 ？
	commands := make([]byte, 5+len(arg))
	commands[4] = command
	copy(commands[5:], arg)
	if err := c.WritePacket(commands); err != nil {
		glog.Errorf("write command failed,%v", err)
		return err
	}
	return nil
}

// Prepare mysql 预编译功能
func (c *Conn) Prepare(query string) (driver.Stmt, error) {
	comd := make([]byte, 0)
	comd = append(comd, constant.ComStmtPrepare)
	comd = append(comd, []byte(query)...)
	if err := c.WriteCommand(comd); err != nil {
		return nil, err
	}
	stmt := new(GdbcStmt)
	stmt.c = c
	if stmtOK, err := stmt.HandlerStmtOK(); err != nil {
		return nil, err
	} else {
		stmt.stmtPack = stmtOK
	}
	return stmt, nil
}

// Close 关闭mysql conn
func (c *Conn) Close() error {
	if c.netConn != nil {
		c.netConn.Close()
	}
	return nil
}

//Begin Tx start
func (c *Conn) Begin() (driver.Tx, error) {
	if err := c.WriteCommandStr(constant.ComQuery, "BEGIN"); err != nil {
		glog.Error(err)
		return nil, err
	}
	tx := GdbcTx{c}
	return tx, nil
}

//Ping method is check this conn is Available
func (c Conn) Ping(ctx context.Context) error {
	com := []byte{constant.ComPing}
	if err := c.WriteCommand(com); err != nil {
		glog.Errorf("send Ping command error:%v", err)
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
