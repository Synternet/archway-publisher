package types

import "google.golang.org/protobuf/reflect/protoreflect"

type Transaction struct {
	Nonce    string `json:"nonce"`
	Raw      string `json:"raw"`
	Code     uint32 `json:"code"`
	TxID     string `json:"tx_id"`
	Tx       any    `json:"tx"`
	TxResult any    `json:"tx_result"`
	Metadata any    `json:"metadata"`
}

type Block struct {
	Nonce string `json:"nonce"`
	Block any    `json:"block"`
}

func (*Mempool) ProtoReflect() protoreflect.Message { return nil }

type Mempool struct {
	Transactions []*Transaction `json:"txs"`
}
