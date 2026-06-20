package http

import (
	"context"
	"io"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/dangerous-drive-guard/backend/internal/blockchain/service"
	"github.com/dangerous-drive-guard/backend/pkg/response"
)

type BlockchainHandler struct {
	bcService *service.BlockchainService
}

func NewBlockchainHandler(svc *service.BlockchainService) *BlockchainHandler {
	return &BlockchainHandler{bcService: svc}
}

func (h *BlockchainHandler) RegisterRoutes(r *app.RouterGroup, authMiddleware app.HandlerFunc) {
	blockchain := r.Group("/blockchain", authMiddleware)
	{
		blockchain.POST("/data", h.UploadData)
		blockchain.GET("/data/:key", h.QueryData)
		blockchain.GET("/data/:key/history", h.GetHistory)
		blockchain.GET("/block/:height", h.QueryBlock)
		blockchain.GET("/transaction/:tx_hash", h.QueryTransaction)
		blockchain.GET("/info", h.GetChainInfo)
		blockchain.POST("/verify", h.VerifyEvidence)
	}
}

type UploadDataRequest struct {
	DataType   string      `json:"data_type" binding:"required"`
	BusinessID int64       `json:"business_id" binding:"required"`
	BusinessNO string      `json:"business_no"`
	Data       interface{} `json:"data" binding:"required"`
}

func (h *BlockchainHandler) UploadData(c context.Context, ctx *app.RequestContext) {
	var req UploadDataRequest
	if err := ctx.BindAndValidate(&req); err != nil {
		response.BadRequest(ctx, err.Error())
		return
	}

	userID, _ := ctx.Get("user_id")
	submittedBy, _ := userID.(int64)

	record, err := h.bcService.SaveData(c, req.DataType, req.BusinessID, req.BusinessNO, req.Data)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	record.SubmittedBy = submittedBy

	response.Success(ctx, map[string]interface{}{
		"id":           record.ID,
		"tx_hash":      record.TxHash,
		"block_height": record.BlockNo,
		"data_hash":    record.DataHash,
		"timestamp":    record.SubmitTime,
		"status":       record.ChainStatus,
	})
}

func (h *BlockchainHandler) QueryData(c context.Context, ctx *app.RequestContext) {
	key := ctx.Param("key")
	if key == "" {
		response.BadRequest(ctx, "key is required")
		return
	}

	record, err := h.bcService.QueryData(c, key)
	if err != nil {
		if err.Error() == "evidence not found" {
			response.NotFound(ctx, err.Error())
			return
		}
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, record)
}

func (h *BlockchainHandler) GetHistory(c context.Context, ctx *app.RequestContext) {
	key := ctx.Param("key")
	if key == "" {
		response.BadRequest(ctx, "key is required")
		return
	}

	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "10"))

	result, err := h.bcService.GetHistory(c, key, page, pageSize)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Page(ctx, result.List, result.Total, result.Page, result.PageSize)
}

func (h *BlockchainHandler) QueryBlock(c context.Context, ctx *app.RequestContext) {
	heightStr := ctx.Param("height")
	height, err := strconv.ParseInt(heightStr, 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid block height")
		return
	}

	block, err := h.bcService.QueryBlock(c, height)
	if err != nil {
		if err.Error() == "block not found" {
			response.NotFound(ctx, err.Error())
			return
		}
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, block)
}

func (h *BlockchainHandler) QueryTransaction(c context.Context, ctx *app.RequestContext) {
	txHash := ctx.Param("tx_hash")
	if txHash == "" {
		response.BadRequest(ctx, "transaction hash is required")
		return
	}

	tx, err := h.bcService.QueryTransaction(c, txHash)
	if err != nil {
		if err.Error() == "transaction not found" {
			response.NotFound(ctx, err.Error())
			return
		}
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, tx)
}

func (h *BlockchainHandler) GetChainInfo(c context.Context, ctx *app.RequestContext) {
	info, err := h.bcService.GetChainInfo(c)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, info)
}

type VerifyEvidenceRequest struct {
	Hash string      `json:"hash"`
	Data interface{} `json:"data"`
}

func (h *BlockchainHandler) VerifyEvidence(c context.Context, ctx *app.RequestContext) {
	hash := ctx.Query("hash")
	var data interface{}

	if hash == "" {
		var req VerifyEvidenceRequest
		body, err := io.ReadAll(ctx.Request.Body())
		if err == nil {
			ctx.Request.SetBody(body)
			if err := ctx.BindAndValidate(&req); err == nil {
				hash = req.Hash
				data = req.Data
			}
		}
	}

	if hash == "" {
		response.BadRequest(ctx, "hash is required")
		return
	}

	if data == nil {
		var req VerifyEvidenceRequest
		if err := ctx.BindAndValidate(&req); err == nil {
			data = req.Data
		}
	}

	if data == nil {
		response.BadRequest(ctx, "data is required for verification")
		return
	}

	valid, err := h.bcService.VerifyHash(c, hash, data)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	response.Success(ctx, map[string]interface{}{
		"valid": valid,
		"hash":  hash,
	})
}
