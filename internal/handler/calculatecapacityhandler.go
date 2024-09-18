package handler

import (
	"net/http"

	"power/internal/logic"
	"power/internal/svc"
	"power/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func calculateCapacityHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CapacityConfigRequest	
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewCalculateCapacityLogic(r.Context(), svcCtx)
		resp, err := l.CalculateCapacity(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
