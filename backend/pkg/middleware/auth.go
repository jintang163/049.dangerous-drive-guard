package middleware

import (
	"context"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/dangerous-drive-guard/backend/pkg/config"
	"github.com/dangerous-drive-guard/backend/pkg/response"
)

type UserClaims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	OrgID    int64  `json:"org_id"`
	jwt.RegisteredClaims
}

func TraceID() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		traceID := ctx.Request.Header.Get("X-Trace-ID")
		if traceID == "" {
			traceID = uuid.New().String()
		}
		ctx.Set("X-Trace-ID", traceID)
		ctx.Header("X-Trace-ID", traceID)
		c = context.WithValue(c, "trace_id", traceID)
		ctx.Next(c)
	}
}

func CORS() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		ctx.Header("Access-Control-Allow-Origin", "*")
		ctx.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		ctx.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, X-Trace-ID")
		ctx.Header("Access-Control-Expose-Headers", "Content-Disposition, X-Trace-ID")
		ctx.Header("Access-Control-Max-Age", "86400")
		if string(ctx.Method()) == "OPTIONS" {
			ctx.AbortWithStatus(204)
			return
		}
		ctx.Next(c)
	}
}

func JWTAuth() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		authHeader := ctx.Request.Header.Get("Authorization")
		if authHeader == "" {
			response.Unauthorized(ctx, "missing authorization header")
			ctx.Abort()
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Unauthorized(ctx, "invalid authorization format")
			ctx.Abort()
			return
		}
		tokenStr := parts[1]
		claims := &UserClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			return []byte(config.Global.JWT.Secret), nil
		})
		if err != nil || !token.Valid {
			response.Unauthorized(ctx, "invalid or expired token")
			ctx.Abort()
			return
		}
		ctx.Set("user_id", claims.UserID)
		ctx.Set("username", claims.Username)
		ctx.Set("role", claims.Role)
		ctx.Set("org_id", claims.OrgID)
		c = context.WithValue(c, "user_claims", claims)
		ctx.Next(c)
	}
}

func RoleAuth(roles ...string) app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		userRole, _ := ctx.Get("role")
		roleStr, ok := userRole.(string)
		if !ok {
			response.Forbidden(ctx, "role not found")
			ctx.Abort()
			return
		}
		for _, r := range roles {
			if r == roleStr {
				ctx.Next(c)
				return
			}
		}
		response.Forbidden(ctx, "insufficient permissions")
		ctx.Abort()
	}
}

func RateLimit(max int64, window time.Duration) app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		key := "ratelimit:" + ctx.ClientIP()
		redis := ctx.MustGet("redis").(interface{ Get(context.Context, string) interface{ Int64() (int64, error) } })
		_ = key
		_ = max
		_ = window
		ctx.Next(c)
	}
}

func GenerateToken(userID int64, username, role string, orgID int64) (string, error) {
	claims := UserClaims{
		UserID:   userID,
		Username: username,
		Role:     role,
		OrgID:    orgID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(config.Global.JWT.ExpireHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "ddg-system",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.Global.JWT.Secret))
}
