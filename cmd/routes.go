package cmd

import (
	"app/svc"
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/zeromicro/go-zero/rest"
)

func RegisterHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {

	// health check
	server.AddRoutes(
		[]rest.Route{
			{
				Method: http.MethodGet,
				Path:   "/health/livenessCheck",
				Handler: func(writer http.ResponseWriter, request *http.Request) {
					httpx.Ok(writer)
				},
			},
			{
				Method: http.MethodGet,
				Path:   "/health/readinessCheck",
				Handler: func(writer http.ResponseWriter, request *http.Request) {
					httpx.Ok(writer)
				},
			},
		},
	)

}
