package model

import (
	"time"
)

type ChainStatus string

const (
	ChainStatusSubmitting ChainStatus = "submitting"
	ChainStatusConfirmed  ChainStatus = "confirmed"
	ChainStatusFailed     ChainStatus = "failed"
)

type EvidenceRecord struct {
	BaseModel
	TxHash       string      `gorm:"type:varchar(128);uniqueIndex;not null" json:"tx_hash"`
	BlockNo      int64       `gorm:"index" json:"block_no"`
	DataType     string      `gorm:"type:varchar(32);index;not null" json:"data_type"`
	BusinessID   int64       `gorm:"index;not null" json:"business_id"`
	BusinessNO   string      `gorm:"type:varchar(64)" json:"business_no"`
	DataHash     string      `gorm:"type:varchar(128);not null" json:"data_hash"`
	PreviousHash string      `gorm:"type:varchar(128)" json:"previous_hash"`
	Payload      JSON        `gorm:"type:json" json:"payload"`
	SubmittedBy  int64       `gorm:"index" json:"submitted_by"`
	SubmitTime   time.Time   `gorm:"not null" json:"submit_time"`
	ConfirmedTime *time.Time `json:"confirmed_time"`
	ChainStatus  ChainStatus `gorm:"type:varchar(20);index;not null" json:"chain_status"`
	ErrorMsg     string      `gorm:"type:varchar(512)" json:"error_msg"`
	NodeInfo     string      `gorm:"type:varchar(256)" json:"node_info"`
}

func (EvidenceRecord) TableName() string {
	return "blockchain_records"
}
