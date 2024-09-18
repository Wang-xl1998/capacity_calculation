package handler

import (
	"net/http"

	"power/internal/logic"
	"power/internal/svc"
	"power/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func queryDataHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.QueryRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		logx.Infof("Parsed request: startTime=%s, endTime=%s, company=%s", req.StartTime, req.EndTime, req.Company)

		l := logic.NewQueryDataLogic(r.Context(), svcCtx)
		resp, err := l.QueryData(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
