// Package money defines the currency scope of an analysis request. Amounts in
// different currencies are never added together without an explicit FX rate.
package money

import (
	"context"
	"github.com/gin-gonic/gin"
	"strings"
)

type currencyKey struct{}

func Currency(ctx context.Context) string {
	if v, ok := ctx.Value(currencyKey{}).(string); ok && v != "" {
		return v
	}
	return "CNY"
}
func WithCurrency(ctx context.Context, c string) context.Context {
	return context.WithValue(ctx, currencyKey{}, c)
}
func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		v := strings.ToUpper(c.DefaultQuery("currency", "CNY"))
		if v != "CNY" && v != "USD" {
			c.AbortWithStatusJSON(400, gin.H{"code": 400, "message": "currency must be CNY or USD"})
			return
		}
		c.Request = c.Request.WithContext(WithCurrency(c.Request.Context(), v))
		c.Next()
	}
}
