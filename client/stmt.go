package client

import (
	"database/sql/driver"
	"encoding/binary"
	"fmt"
	"gdbc/constant"
	"gdbc/packets"
	"gdbc/utils"
	"github.com/golang/glog"
)

//GdbcStmt is stmt
type GdbcStmt struct {
	c        *Conn
	stmtPack *StmtPrepareOk
}

//Close function is stmt close
func (st GdbcStmt) Close() error {
	command := make([]byte, 5)
	command[0] = constant.ComStmtClose
	binary.LittleEndian.PutUint32(command[1:], st.stmtPack.statementID)
	st.c.WriteCommand(command)
	return nil
}

//NumInput is return params number
func (st GdbcStmt) NumInput() int {
	return int(st.stmtPack.numParams)
}

//StmtExecuteResponse is service send COM_STMT_EXECUTE command result
// type StmtExecuteResponse struct {
// 	columnCount      uint64
// 	columnDefinition []ColumnDef41Packets
// 	results          []BinaryResultsetRow
// }

// 文档中的num-params 这个参数就是 numParams
// Exec : https://dev.mysql.com/doc/internals/en/com-stmt-execute.html
func (st GdbcStmt) Exec(args []driver.Value) (driver.Result, error) {
	if int(st.stmtPack.numParams) != len(args) {
		return nil, fmt.Errorf("argument mismatch, need %d but got %d", st.stmtPack.numParams, len(args))
	}
	command := execParamsConstruct(args, st.stmtPack.numParams, st.stmtPack.statementID)
	if err := st.c.WriteCommand(command); err != nil {
		glog.Errorf("send COM_STMT_EXECUTE command error:%+v", err)
		return nil, err
	}
	//if err := st.c.Payload.ReadPayload(st.c.netConn); err != nil {
	//	glog.Errorf("read COM_STMT_EXECUTE result error:%+v", err)
	//	return nil, err
	//}
	//a OK_Packet
	//a ERR_Packet
	//or a resultset: Binary Protocol Resultset
	// binaryResult := BinaryResultset{}
	// if okPackt := packets.OkPacketHandler(st.c.Payload.Payload, st.c.Payload.Capability); okPackt.IsOk() {
	// 	binaryResult.result.affectedRows = okPackt.AffectedRows
	// 	binaryResult.result.lastInsertID = okPackt.LastInsertId
	// 	return binaryResult.result, nil
	// }
	// if errPackt := packets.ErrPacketHandler(st.c.Payload.Payload, st.c.Payload.Capability); errPackt.IsErr() {
	// 	glog.Errorf("COM_STMT_EXECUTE command,service return err packet:%s", errPackt.String())
	// 	return nil, errors.New(errPackt.String())
	// }
	// stmtExecRsp := new(StmtExecuteResponse)
	// stmtExecRsp.columnCount, _ = utils.LenencIntDecode(st.c.Payload.Payload)
	// if columnDefinition, err := HandlerColumnDef41s(st.c); err != nil {
	// 	return nil, err
	// } else {
	// 	stmtExecRsp.columnDefinition = columnDefinition
	// }
	// if results, err := HandlerResultSetRowS(st.c, stmtExecRsp.columnCount); err != nil {
	// 	return nil, err
	// } else {
	// 	stmtExecRsp.results = results
	// }
	if binaryResult, err := HandlerBinaryResultSet(st.c); err != nil {
		glog.Error(err)
		return nil, err
	} else if binaryResult.result != nil {
		return binaryResult.result, nil
	}
	return nil, nil
}

//Query is select command user,other command call exec method
func (st GdbcStmt) Query(args []driver.Value) (driver.Rows, error) {
	if int(st.stmtPack.numParams) != len(args) {
		return nil, fmt.Errorf("argument mismatch, need %d but got %d", st.stmtPack.numParams, len(args))
	}
	command := execParamsConstruct(args, st.stmtPack.numParams, st.stmtPack.statementID)
	if err := st.c.WriteCommand(command); err != nil {
		glog.Errorf("send COM_STMT_EXECUTE command error:%+v", err)
		return nil, err
	}
	rows := new(GdbcRows)
	rows.c = st.c
	rows.protocolType = constant.BinaryProtocol
	if err := rows.HandlerRows(); err != nil {
		glog.Error(err)
		return nil, err
	}
	return rows, nil
}

//StmtPrepareOk is  COM_STMT_PREPARE command successed,server send result
type StmtPrepareOk struct {
	status       uint8                //[00] OK
	statementID  uint32               //statement_id (4) -- statement-id
	numColumns   uint16               //num_columns (2) -- number of columns
	numParams    uint16               //num_params (2) -- number of params
	warningCount uint16               // warning_count (2) -- number of warnings
	parameter    []ColumnDef41Packets //If num_params > 0 more packets will follow:
	column       []ColumnDef41Packets //If num_columns > 0 more packets will follow:
}

//HandlerStmtOK handler OK packtes
func (st GdbcStmt) HandlerStmtOK() (*StmtPrepareOk, error) {
	if err := st.c.Payload.ReadPayload(st.c.netConn); err != nil {
		return nil, err
	}
	stmtOK := new(StmtPrepareOk)
	payload := st.c.Payload.Payload
	index := 0
	stmtOK.status = payload[index]
	index++
	stmtOK.statementID = binary.LittleEndian.Uint32(payload[index:])
	index += 4
	stmtOK.numColumns = binary.LittleEndian.Uint16(payload[index:])
	index += 2
	stmtOK.numParams = binary.LittleEndian.Uint16(payload[index:])
	index += 2
	stmtOK.warningCount = binary.LittleEndian.Uint16(payload[index:])
	if stmtOK.numParams > 0 {
		params := make([]ColumnDef41Packets, 0)
		for {
			if err := st.c.Payload.ReadPayload(st.c.netConn); err != nil {
				return nil, err
			}
			payload = st.c.Payload.Payload
			if eofPack := packets.EOFPacketHandler(payload, st.c.Payload.Capability); eofPack.IsEOF() {
				break
			}
			column := HandlerColumnDef41(payload)
			params = append(params, column)
		}
		stmtOK.parameter = params
	}

	if stmtOK.numColumns > 0 {
		params := make([]ColumnDef41Packets, 0)
		for {
			if err := st.c.Payload.ReadPayload(st.c.netConn); err != nil {
				return nil, err
			}
			payload = st.c.Payload.Payload
			if eofPack := packets.EOFPacketHandler(payload, st.c.Payload.Capability); eofPack.IsEOF() {
				break
			}
			column := HandlerColumnDef41(payload)
			params = append(params, column)
		}
		stmtOK.column = params
	}
	return stmtOK, nil
}

func execParamsConstruct(args []driver.Value, numParams uint16, statementID uint32) (command []byte) {
	// 1              [17] COM_STMT_EXECUTE
	// 4              stmt-id
	// 1              flags 0x00
	// 4              iteration-count ,The iteration-count is always 1.
	minPacketsLength := 1 + 4 + 1 + 4
	command = make([]byte, minPacketsLength)
	command[0] = constant.ComStmtExecute
	command[1] = uint8(statementID)
	command[2] = uint8(statementID >> 8)
	command[3] = uint8(statementID >> 16)
	command[4] = uint8(statementID >> 24)
	command[5] = 0x00
	command[6], command[7], command[8], command[9] = 0x01, 0x00, 0x00, 0x00
	num_params := numParams
	if num_params > 0 {
		// 除3 用右移运算符,这里还有一个offset 计算,(num-fields + 7 + offset) / 8 ,在这个命令时offset=0
		NULLBitmap := make([]byte, (num_params+7)>>3)
		var newParamsBoundFlag byte = 0x01
		//type of each parameter, length: num-params * 2
		paramsType := make([]byte, num_params<<1)
		paramsValue := make([][]byte, num_params)
		for i, arg := range args {
			if arg == nil {
				NULLBitmap[i/8] |= (1 << (uint(i) % 8))
				NULLBitmap[i<<1] = constant.MysqlTypeNull
				continue
			}
			ptype, value := utils.DriverValueHandler(arg, constant.BinaryProtocol)
			paramsType[i<<1] = ptype[0]
			paramsType[(i<<1)+1] = ptype[1]
			paramsValue[i] = value
		}
		command = append(command, NULLBitmap...)
		command = append(command, newParamsBoundFlag)
		command = append(command, paramsType...)
		for _, value := range paramsValue {
			command = append(command, value...)
		}
	}
	return
}
