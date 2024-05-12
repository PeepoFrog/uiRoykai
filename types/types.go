package types

type InterxInfo struct {
	CatchingUp        bool   `json:"catching_up"`
	ChainID           string `json:"chain_id"`
	FaucetAddr        string `json:"faucet_addr"`
	GenesisChecksum   string `json:"genesis_checksum"`
	KiraAddr          string `json:"kira_addr"`
	KiraPubKey        string `json:"kira_pub_key"`
	LatestBlockHeight string `json:"latest_block_height"`
	Moniker           string `json:"moniker"`
	Node              Node   `json:"node"`
	PubKey            PubKey `json:"pub_key"`
	Version           string `json:"version"`
}

type Node struct {
	NodeType        string `json:"node_type"`
	SeedNodeID      string `json:"seed_node_id"`
	SentryNodeID    string `json:"sentry_node_id"`
	SnapshotNodeID  string `json:"snapshot_node_id"`
	ValidatorNodeID string `json:"validator_node_id"`
}

type PubKey struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type NodeInfo struct {
	Channels        string          `json:"channels"`
	ID              string          `json:"id"`
	ListenAddr      string          `json:"listen_addr"`
	Moniker         string          `json:"moniker"`
	Network         string          `json:"network"`
	Other           Other           `json:"other"`
	ProtocolVersion ProtocolVersion `json:"protocol_version"`
	Version         string          `json:"version"`
}

type Other struct {
	RpcAddress string `json:"rpc_address"`
	TxIndex    string `json:"tx_index"`
}

type ProtocolVersion struct {
	App   string `json:"app"`
	Block string `json:"block"`
	P2p   string `json:"p2p"`
}

type SyncInfo struct {
	EarliestAppHash     string `json:"earliest_app_hash"`
	EarliestBlockHash   string `json:"earliest_block_hash"`
	EarliestBlockHeight string `json:"earliest_block_height"`
	EarliestBlockTime   string `json:"earliest_block_time"`
	LatestAppHash       string `json:"latest_app_hash"`
	LatestBlockHash     string `json:"latest_block_hash"`
	LatestBlockHeight   string `json:"latest_block_height"`
	LatestBlockTime     string `json:"latest_block_time"`
}

type ValidatorInfo struct {
	Address     string `json:"address"`
	PubKey      PubKey `json:"pub_key"`
	VotingPower string `json:"voting_power"`
}

type Info struct {
	ID            string        `json:"id"`
	InterxInfo    InterxInfo    `json:"interx_info"`
	NodeInfo      NodeInfo      `json:"node_info"`
	SyncInfo      SyncInfo      `json:"sync_info"`
	ValidatorInfo ValidatorInfo `json:"validator_info"`
}
