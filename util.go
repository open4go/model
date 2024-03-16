package model

import (
	"context"
	"github.com/open4go/log"
)

// GetValueFromCtx 从context中读取值
func GetValueFromCtx(ctx context.Context, key string) string {
	// Retrieve the value from the context
	value := ctx.Value(key)

	// Check if the value is of the expected type
	if str, ok := value.(string); ok {
		log.Log().WithField("key", key).WithField("value", str).
			Debug("value retrieved from context")
		return str
	} else {
		log.Log().WithField("key", key).Warning("Value not found or not of type string")
		return ""
	}
}
