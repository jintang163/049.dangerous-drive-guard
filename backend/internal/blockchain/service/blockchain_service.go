package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dangerous-drive-guard/backend/internal/common/model"
	"github.com/dangerous-drive-guard/backend/pkg/config"
	"github.com/dangerous-drive-guard/backend/pkg/database"
	"github.com/dangerous-drive-guard/backend/pkg/logger"
	"gorm.io/gorm"
)

type BlockchainService struct {
	db          *gorm.DB
	cfg         *config.BlockchainConfig
	chainHeight int64
}

type BlockInfo struct {
	Height     int64     `json:"height"`
	Hash       string    `json:"hash"`
	PrevHash   string    `json:"prev_hash"`
	TxCount    int       `json:"tx_count"`
	Timestamp  time.Time `json:"timestamp"`
	Nonce      string    `json:"nonce"`
	Difficulty string    `json:"difficulty"`
}

type TransactionInfo struct {
	TxHash      string      `json:"tx_hash"`
	BlockHeight int64       `json:"block_height"`
	Timestamp   string      `json:"timestamp"`
	Sender      string      `json:"sender"`
	Recipient   string      `json:"recipient"`
	Payload     interface{} `json:"payload"`
}

type ChainInfo struct {
	ChannelID   string `json:"channel_id"`
	Chaincode   string `json:"chaincode"`
	Status      string `json:"status"`
	BlockHeight int64  `json:"block_height"`
	PeerCount   int    `json:"peer_count"`
}

type EvidencePage struct {
	List     []*model.EvidenceRecord `json:"list"`
	Total    int64                   `json:"total"`
	Page     int                     `json:"page"`
	PageSize int                     `json:"page_size"`
}

func NewBlockchainService(cfg *config.Config) *BlockchainService {
	var maxBlockNo int64
	database.GetDB().Model(&model.EvidenceRecord{}).Select("COALESCE(MAX(block_no), 0)").Scan(&maxBlockNo)

	return &BlockchainService{
		db:          database.GetDB(),
		cfg:         &cfg.Blockchain,
		chainHeight: maxBlockNo,
	}
}

func calculateHash(data []byte) string {
	h := sha256.New()
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

func calculateChainHash(prevHash, dataHash string, timestamp int64) string {
	combined := fmt.Sprintf("%s%s%d", prevHash, dataHash, timestamp)
	return calculateHash([]byte(combined))
}

func (s *BlockchainService) SaveData(ctx context.Context, dataType string, businessID int64, businessNO string, data interface{}) (*model.EvidenceRecord, error) {
	if dataType == "" {
		return nil, fmt.Errorf("data type is required")
	}
	if businessID <= 0 {
		return nil, fmt.Errorf("business id is required")
	}
	if data == nil {
		return nil, fmt.Errorf("data is required")
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		logger.Sugar.Errorf("marshal data error: %v", err)
		return nil, fmt.Errorf("marshal data error: %w", err)
	}

	dataHash := calculateHash(jsonBytes)

	var lastRecord model.EvidenceRecord
	prevHash := ""
	err = s.db.WithContext(ctx).Order("id DESC").First(&lastRecord).Error
	if err == nil && lastRecord.ID > 0 {
		prevHash = lastRecord.DataHash
	}

	now := time.Now()
	txHash := calculateChainHash(prevHash, dataHash, now.UnixNano())

	s.chainHeight++

	payloadJSON, _ := json.Marshal(map[string]interface{}{
		"data":      data,
		"data_type": dataType,
		"timestamp": now,
	})

	record := &model.EvidenceRecord{
		TxHash:       txHash,
		BlockNo:      s.chainHeight,
		DataType:     dataType,
		BusinessID:   businessID,
		BusinessNO:   businessNO,
		DataHash:     dataHash,
		PreviousHash: prevHash,
		Payload:      model.JSON(payloadJSON),
		SubmitTime:   now,
		ChainStatus:  model.ChainStatusConfirmed,
		ConfirmedTime: &now,
		NodeInfo:     "local-node-1",
	}

	if err := s.db.WithContext(ctx).Create(record).Error; err != nil {
		logger.Sugar.Errorf("create evidence record error: %v", err)
		return nil, fmt.Errorf("save evidence error: %w", err)
	}

	logger.Sugar.Infof("blockchain data saved: type=%s, business_id=%d, tx_hash=%s", dataType, businessID, txHash)
	return record, nil
}

func (s *BlockchainService) QueryData(ctx context.Context, key string) (*model.EvidenceRecord, error) {
	if key == "" {
		return nil, fmt.Errorf("key is required")
	}

	var record model.EvidenceRecord
	err := s.db.WithContext(ctx).Where("tx_hash = ? OR data_hash = ?", key, key).First(&record).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("evidence not found")
		}
		logger.Sugar.Errorf("query evidence error: %v", err)
		return nil, fmt.Errorf("query evidence error: %w", err)
	}

	return &record, nil
}

func (s *BlockchainService) GetHistory(ctx context.Context, key string, page, pageSize int) (*EvidencePage, error) {
	if key == "" {
		return nil, fmt.Errorf("key is required")
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}

	var records []*model.EvidenceRecord
	var total int64

	query := s.db.WithContext(ctx).Model(&model.EvidenceRecord{}).
		Where("data_type = ? OR business_no = ?", key, key)

	if err := query.Count(&total).Error; err != nil {
		logger.Sugar.Errorf("count evidence history error: %v", err)
		return nil, fmt.Errorf("count error: %w", err)
	}

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&records).Error; err != nil {
		logger.Sugar.Errorf("query evidence history error: %v", err)
		return nil, fmt.Errorf("query error: %w", err)
	}

	return &EvidencePage{
		List:     records,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *BlockchainService) QueryBlock(ctx context.Context, height int64) (*BlockInfo, error) {
	if height <= 0 {
		return nil, fmt.Errorf("invalid block height")
	}

	var records []model.EvidenceRecord
	err := s.db.WithContext(ctx).Where("block_no = ?", height).Find(&records).Error
	if err != nil {
		logger.Sugar.Errorf("query block error: %v", err)
		return nil, fmt.Errorf("query block error: %w", err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("block not found")
	}

	var prevRecord model.EvidenceRecord
	prevHash := ""
	s.db.WithContext(ctx).Where("block_no = ?", height-1).Order("id DESC").First(&prevRecord)
	if prevRecord.ID > 0 {
		prevHash = prevRecord.DataHash
	}

	firstRecord := records[0]
	blockHash := calculateChainHash(prevHash, firstRecord.DataHash, firstRecord.SubmitTime.Unix())

	return &BlockInfo{
		Height:     height,
		Hash:       blockHash,
		PrevHash:   prevHash,
		TxCount:    len(records),
		Timestamp:  firstRecord.SubmitTime,
		Nonce:      fmt.Sprintf("%d", firstRecord.ID),
		Difficulty: "1",
	}, nil
}

func (s *BlockchainService) QueryTransaction(ctx context.Context, txHash string) (*TransactionInfo, error) {
	if txHash == "" {
		return nil, fmt.Errorf("transaction hash is required")
	}

	var record model.EvidenceRecord
	err := s.db.WithContext(ctx).Where("tx_hash = ?", txHash).First(&record).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("transaction not found")
		}
		logger.Sugar.Errorf("query transaction error: %v", err)
		return nil, fmt.Errorf("query transaction error: %w", err)
	}

	var payload interface{}
	if len(record.Payload) > 0 {
		json.Unmarshal(record.Payload, &payload)
	}

	return &TransactionInfo{
		TxHash:      record.TxHash,
		BlockHeight: record.BlockNo,
		Timestamp:   record.SubmitTime.Format(time.RFC3339),
		Sender:      fmt.Sprintf("user_%d", record.SubmittedBy),
		Recipient:   record.DataType,
		Payload:     payload,
	}, nil
}

func (s *BlockchainService) GetChainInfo(ctx context.Context) (*ChainInfo, error) {
	var maxBlockNo int64
	var totalRecords int64

	s.db.WithContext(ctx).Model(&model.EvidenceRecord{}).
		Select("COALESCE(MAX(block_no), 0)").Scan(&maxBlockNo)
	s.db.WithContext(ctx).Model(&model.EvidenceRecord{}).Count(&totalRecords)

	status := "active"
	if maxBlockNo == 0 {
		status = "idle"
	}

	return &ChainInfo{
		ChannelID:   s.cfg.Fabric.ChannelID,
		Chaincode:   s.cfg.Fabric.ChaincodeName,
		Status:      status,
		BlockHeight: maxBlockNo,
		PeerCount:   1,
	}, nil
}

func (s *BlockchainService) VerifyHash(ctx context.Context, hash string, data interface{}) (bool, error) {
	if hash == "" {
		return false, fmt.Errorf("hash is required")
	}
	if data == nil {
		return false, fmt.Errorf("data is required")
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		logger.Sugar.Errorf("marshal data for verify error: %v", err)
		return false, fmt.Errorf("marshal data error: %w", err)
	}

	calculatedHash := calculateHash(jsonBytes)

	return calculatedHash == hash, nil
}
