package bind

import (
	"live-interact-engine/services/api-service/internal/utils/reskit/response"
	"live-interact-engine/services/api-service/internal/utils/validator"

	"github.com/gin-gonic/gin"
)

// BindingRegularAndResponse 绑定请求体中的 JSON、查询参数和 URI 参数到 req
// 如果绑定失败，自动返回参数错误响应，并返回错误
func BindingRegularAndResponse[T any](ctx *gin.Context, req *T) error {
	_ = ctx.ShouldBind(req)
	_ = ctx.ShouldBindQuery(req)
	_ = ctx.ShouldBindUri(req)

	if err := validator.ValidateStruct(req); err != nil {
		response.InvalidParams(ctx, err)
		return err
	}

	return nil
}
