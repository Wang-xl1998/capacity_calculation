package handler

import (
	"net/http"

	"power/internal/logic"
	"power/internal/svc"
	"power/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func uploadFileHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 创建一个 UploadRequest 结构体实例
		req := &types.UploadRequest{
			Company: r.FormValue("company"), // 直接从请求中获取 company 字段
		}

		// 传递到逻辑层并处理
		l := logic.NewUploadFileLogic(r.Context(), ctx, r)
		resp, err := l.UploadFile(req)
		if err != nil {
			httpx.Error(w, err)
		} else {
			httpx.OkJson(w, resp)
		}
	}
}
