package endpoints

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/nogen-app/prik"
)

type Result struct {
	Status int `json:"status"`
	Body any `json:"body"`
}

type Endpoint struct {
	Method string
	Path string
	Handle func(*prik.Context, echo.Context) *Result
}

type Route func(Endpoint) echo.HandlerFunc
type HandlerFunc[T any] func(*prik.Context, *T) *Result

func CreateEndpoint[T any](
	method string,
	path string,
	handlerFunc HandlerFunc[T],
) Endpoint {
	return Endpoint{
		Method: method,
		Path: path,
		Handle: func(ctx *prik.Context, c echo.Context) *Result {
			var data T

			if err := c.Bind(data); err != nil {
				res := Result{Status: http.StatusBadRequest, Body: err.Error()}
				return &res
			}

			if err := c.Validate(data); err != nil {
				res := Result{Status: http.StatusBadRequest, Body: err.Error()}
				return &res
			}

			return handlerFunc(ctx, &data)
		},
	}
}
	
func CreateEndpoints(context *prik.Context, endpoints []Endpoint, server *echo.Echo) {
	route := createRoute(context)

	for _, e := range endpoints {
		switch e.Method {
		case http.MethodGet:
			server.GET(e.Path, route(e))
		case http.MethodPost:
			server.POST(e.Path, route(e))
		case http.MethodPut:
			server.PUT(e.Path, route(e))
		case http.MethodDelete:
			server.DELETE(e.Path, route(e))
		case http.MethodPatch:
			server.PATCH(e.Path, route(e))
		default:
			panic("Invalid method")
		}
	}
}

func createRoute(context *prik.Context) Route {
	return func(e Endpoint) echo.HandlerFunc {
		return func(c echo.Context) error {
			res := e.Handle(context, c)
			return c.JSON(200, res)
		}
	}
}

