package eth

const (
	eth62 = 62
	eth63 = 63
)

//协议名
var ProtocolName = "eth"

// Supported versions of the eth protocol (first is primary).
var ProtocolVersions = []uint{eth63, eth62}

// 每个协议支持的message个数 去上面的 版本对应
var ProtocolLengths = []uint64{17, 8}

// 以太坊的消息命令字
const (
	// Protocol messages belonging to eth/62
	StatusMsg          = 0x00
	NewBlockHashesMsg  = 0x01
	TxMsg              = 0x02
	GetBlockHeadersMsg = 0x03
	BlockHeadersMsg    = 0x04
	GetBlockBodiesMsg  = 0x05
	BlockBodiesMsg     = 0x06
	NewBlockMsg        = 0x07

	// Protocol messages belonging to eth/63
	GetNodeDataMsg = 0x0d
	NodeDataMsg    = 0x0e
	GetReceiptsMsg = 0x0f
	ReceiptsMsg    = 0x10
)
