package response

import (
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	TraceID string      `json:"trace_id,omitempty"`
}

type PageData struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

const (
	CodeSuccess        = 0
	CodeBadRequest     = 40000
	CodeUnauthorized   = 40100
	CodeForbidden      = 40300
	CodeNotFound       = 40400
	CodeTooManyRequests= 42900
	CodeInternalError  = 50000
	CodeServiceDown    = 50300
)

var codeMessage = map[int]string{
	CodeSuccess:         "success",
	CodeBadRequest:      "bad request",
	CodeUnauthorized:    "unauthorized",
	CodeForbidden:       "forbidden",
	CodeNotFound:        "not found",
	CodeTooManyRequests: "too many requests",
	CodeInternalError:   "internal server error",
	CodeServiceDown:     "service unavailable",
}

func Success(c *app.RequestContext, data ...interface{}) {
	r := &Response{
		Code:    CodeSuccess,
		Message: codeMessage[CodeSuccess],
	}
	if len(data) > 0 {
		r.Data = data[0]
	}
	if traceID, ok := c.Get("X-Trace-ID"); ok {
		r.TraceID = traceID.(string)
	}
	c.JSON(consts.StatusOK, r)
}

func Page(c *app.RequestContext, list interface{}, total int64, page, pageSize int) {
	Success(c, PageData{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

func Fail(c *app.RequestContext, code int, messages ...string) {
	msg := codeMessage[code]
	if len(messages) > 0 && messages[0] != "" {
		msg = messages[0]
	}
	r := &Response{
		Code:    code,
		Message: msg,
	}
	if traceID, ok := c.Get("X-Trace-ID"); ok {
		r.TraceID = traceID.(string)
	}
	httpCode := http.StatusOK
	if code >= 50000 {
		httpCode = http.StatusInternalServerError
	} else if code >= 40000 {
		httpCode = code / 100
	}
	c.JSON(httpCode, r)
}

func BadRequest(c *app.RequestContext, messages ...string) {
	Fail(c, CodeBadRequest, messages...)
}

func Unauthorized(c *app.RequestContext, messages ...string) {
	Fail(c, CodeUnauthorized, messages...)
}

func Forbidden(c *app.RequestContext, messages ...string) {
	Fail(c, CodeForbidden, messages...)
}

func NotFound(c *app.RequestContext, messages ...string) {
	Fail(c, CodeNotFound, messages...)
}

func InternalError(c *app.RequestContext, messages ...string) {
	Fail(c, CodeInternalError, messages...)
}
