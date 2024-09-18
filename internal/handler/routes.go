// Code generated by goctl. DO NOT EDIT.
package handler

import (
	"net/http"

	"power/internal/svc"

	"github.com/zeromicro/go-zero/rest"
)

func RegisterHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {
	server.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodPost,
				Path:    "/capacity/",
				Handler: calculateCapacityHandler(serverCtx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/query/",
				Handler: queryDataHandler(serverCtx),
			},
			{
				Method:  http.MethodPost,
				Path:    "/upload/",
				Handler: uploadFileHandler(serverCtx),
			},
		},
	)
}