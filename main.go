package endpoints

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/nogen-app/prik"
)

type Result struct {
	Status int    `json:"status"`
	Body   string `json:"body"`
}

type Endpoint struct {
	Method string
	Path   string
	Handle func(*prik.Context, *http.Request) *Result
}

type Route func(Endpoint) echo.HandlerFunc
type HandlerFunc func(*prik.Context, *http.Request) *Result

func CreateEndpoint(
	method string,
	path string,
	handlerFunc HandlerFunc,
) Endpoint {
	return Endpoint{
		Method: method,
		Path:   path,
		Handle: func(ctx *prik.Context, req *http.Request) *Result {
			return handlerFunc(ctx, req)
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
			req := c.Request()
			res := e.Handle(context, req)
			return c.JSON(200, res)
		}
	}
}

func DecodeJSONBody[T any](r *http.Request) (*T, error) {
	var data T
	decoder := json.NewDecoder(r.Body)

	defer r.Body.Close()

	err := decoder.Decode(&data)
	if err != nil {
		return nil, err
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	verr := validate.Struct(data)
	if verr != nil {
		return nil, verr
	}

	return &data, nil
}
