// create_by : lengrongfu
package constant

// 协议常量
const (
	MaxPayloadLength   = 1<<24 - 1 //当次请求最大数据16M
	MinProtocolVersion = 10        //支持的最小协议
	InitSequenceID     = 0
)

/**
  检验mysql头的常量
*/
const (
	OkHeader           byte = 0x00
	OkPacketMinLength  int  = 7
	EofHeader          byte = 0xfe
	EofPacketMaxLength int  = 9
	ErrHeader          byte = 0xff
)

/**
  标示mysql返回数据是否正确
*/
const (
	OK = iota
	EOF
	ERR
)

/**
  Capability Flags
*/
const (
	ClientLongPassword uint32 = 1 << iota
	ClientFoundRows
	ClientLongFlag
	ClientConnectWithDb
	ClientNoSchema
	ClientCompress
	ClientOdbc
	ClientLocalFiles
	ClientIgnoreSpace
	ClientProtocol41
	ClientInteractive
	ClientSsl
	ClientIgnoreSigpipe
	ClientTransactions
	ClientReserved
	ClientSecureConnection
	ClientMultiStatements
	ClientMultiResults
	ClientPsMultiResults
	ClientPluginAuth
	ClientConnectAttrs
	ClientPluginAuthLenencClientData
	ClientCanHandleExpiredPasswords
	ClientSessionTrack
	ClientDeprecateEOF
)

/**
  command phase
  https://dev.mysql.com/doc/internals/en/command-phase.html
*/
const (
	ComSleep            byte = 0x00
	ComQuit             byte = 0x01
	ComInitDb           byte = 0x02
	ComQuery            byte = 0x03
	ComFieldList        byte = 0x04
	ComCreateDb         byte = 0x05 //Obsolete
	ComDropDb           byte = 0x06 //Obsolete
	ComRefresh          byte = 0x07
	ComShutdown         byte = 0x08
	ComStatistics       byte = 0x09
	ComProcessInfo      byte = 0x0a
	ComConnect          byte = 0x0b
	ComProcessKill      byte = 0x0c
	ComDebug            byte = 0x0d
	ComPing             byte = 0x0e
	ComTime             byte = 0x0f
	ComDelayedInsert    byte = 0x10
	ComChangeUser       byte = 0x11
	ComResetConnection  byte = 0x1f
	ComDaemon           byte = 0x1d
	ComBinlogDump       byte = 0x12
	ComTableDump        byte = 0x13
	ComConnectOut       byte = 0x14
	ComRegisterSlave    byte = 0x15
	ComStmtPrepare      byte = 0x16
	ComStmtExecute      byte = 0x17
	ComStmtSendLongData byte = 0x18
	ComStmtClose        byte = 0x19
	ComStmtReset        byte = 0x1a
	ComSetOption        byte = 0x1b
	ComStmtFetch        byte = 0x1c
	ComBinlogDumpGtid   byte = 0x1e
)

// default const
const (
	AuthName           string = "mysql_native_password"
	DefaultCharset     string = "utf8"
	DefaultCharsetByte uint8  = 33
)

//auth plugin
const (
	MysqlOldPassword            string = "mysql_old_password"
	MysqlNativePassword         string = "mysql_native_password" //sha1 加密
	MysqlClearPassword          string = "mysql_clear_password"
	AuthenticationWindowsClient string = "authentication_windows_client"
	Sha256Password              string = "sha256_password"
)

//packets Status Flags
const (
	ServerStatusInTrans            uint16 = 0x0001 //a transaction is active
	ServerStatusAutocommit         uint16 = 0x0002 //auto-commit is enabled
	ServerMoreResultsExists        uint16 = 0x0008
	ServerStatusNoGoodIndexUsed    uint16 = 0x0010
	ServerStatusNoIndexUsed        uint16 = 0x0020
	ServerStatusCursorExists       uint16 = 0x0040
	ServerStatusLastRowSent        uint16 = 0x0080
	ServerStatusDBDropped          uint16 = 0x0100
	ServerStatusNoBackslashEscapes uint16 = 0x0200

	ServerStatusMetadataChanged uint16 = 0x0400

	ServerQueryWasSlow uint16 = 0x0800

	ServerPsOutParams uint16 = 0x1000

	ServerStatusInTransReadOnly uint16 = 0x2000
	ServerSessionStateChanged   uint16 = 0x4000
)

//Character Set
const ()

//Column Types
const (
	MysqlTypeDecimal    byte = 0x00
	MysqlTypeTiny       byte = 0x01
	MysqlTypeShort      byte = 0x02
	MysqlTypeLong       byte = 0x03
	MysqlTypeFloat      byte = 0x04
	MysqlTypeDouble     byte = 0x05
	MysqlTypeNull       byte = 0x06
	MysqlTypeTimestamp  byte = 0x07
	MysqlTypeLongLong   byte = 0x08
	MysqlTypeInt24      byte = 0x09
	MysqlTypeDate       byte = 0x0a
	MysqlTypeTime       byte = 0x0b
	MysqlTypeDatetime   byte = 0x0c
	MysqlTypeYear       byte = 0x0d
	MysqlTypeNewDate    byte = 0x0e
	MysqlTypeVarchar    byte = 0x0f
	MysqlTypeBit        byte = 0x10
	MysqlTypeTimestamp2 byte = 0x11
	MysqlTypeDatetime2  byte = 0x12
	MysqlTypeTime2      byte = 0x13
	MysqlTypeJSON       byte = 0xf5
	MysqlTypeNewDecimal byte = 0xf6
	MysqlTypeEnum       byte = 0xf7
	MysqlTypeSet        byte = 0xf8
	MysqlTypeTinyBlob   byte = 0xf9
	MysqlTypeMediumBlob byte = 0xfa
	MysqlTypeLongBlog   byte = 0xfb
	MysqlTypeBlob       byte = 0xfc
	MysqlTypeVarString  byte = 0xfd
	MysqlTypeString     byte = 0xfe
	MysqlTypeGeometry   byte = 0xff
)

//字段 flag
const (
	NotNullFlag       uint16 = 0x0001
	PriKeyFlag        uint16 = 0x0002
	UniqueKeyFlag     uint16 = 0x0004
	MultipleKeyFlag   uint16 = 0x0008
	BlobFlag          uint16 = 0x0010
	UnsignedFlag      uint16 = 0x0020
	ZerofillFlag      uint16 = 0x0040
	BinaryFlag        uint16 = 0x0080
	EnumFlag          uint16 = 0x0100
	AutoIncrementFlag uint16 = 0x0200
	TimestampFlag     uint16 = 0x0400
	SetFlag           uint16 = 0x0800
)

//protocol type
const (
	TextProtocol = iota
	BinaryProtocol
)
