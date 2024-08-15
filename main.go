package endpoints

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/nogen-app/prik"
)

type Result struct {
	Status int `json:"status"`
	Body any `json:"body"`
}

type Endpoint struct {
	method string
	path string
	handle func(*prik.Context, echo.Context) *Result
}

type Route func(Endpoint) echo.HandlerFunc
type HandlerFunc[T any] func(*prik.Context, *T) *Result
type PassthroughFunc func(*prik.Context, *http.Request) *Result

func CreateEndpoint[T any](
	method string,
	path string,
	handlerFunc HandlerFunc[T],
) Endpoint {
	return Endpoint{
		method: method,
		path: path,
		handle: func(ctx *prik.Context, c echo.Context) *Result {
			var data T

			if err := c.Bind(&data); err != nil {
				res := Result{Status: http.StatusBadRequest, Body: err.Error()}
				return &res
			}

			validate := validator.New(validator.WithRequiredStructEnabled())

			if err := validate.Struct(data); err != nil {
				res := Result{Status: http.StatusBadRequest, Body: err.Error()}
				return &res
			}

			return handlerFunc(ctx, &data)
		},
	}
}
	
func CreatePassthroughEndpoint(
	method string,
	path string,
	handlerFunc PassthroughFunc,
) Endpoint {
	return Endpoint{
		method: method,
		path: path,
		handle: func(ctx *prik.Context, c echo.Context) *Result {
			req := c.Request()
			return handlerFunc(ctx, req)
		},
	}
}

func CreateEndpoints(context *prik.Context, endpoints []Endpoint, server *echo.Echo) {
	route := createRoute(context)

	for _, e := range endpoints {
		switch e.method {
		case http.MethodGet:
			server.GET(e.path, route(e))
		case http.MethodPost:
			server.POST(e.path, route(e))
		case http.MethodPut:
			server.PUT(e.path, route(e))
		case http.MethodDelete:
			server.DELETE(e.path, route(e))
		case http.MethodPatch:
			server.PATCH(e.path, route(e))
		default:
			panic("Invalid method")
		}
	}
}

func createRoute(context *prik.Context) Route {
	return func(e Endpoint) echo.HandlerFunc {
		return func(c echo.Context) error {
			res := e.handle(context, c)
			return c.JSON(200, res)
		}
	}
}

