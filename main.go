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
	tag string
	json *JSONEndpoint
	streaming *StreamingEndpoint
}

type JSONEndpoint struct {
	method string
	path string
	handle func(*prik.Context, echo.Context) *Result
}

type StreamingEndpoint struct {
	method string
	path string
	handle func(*prik.Context, echo.Context) *http.Response
}

type Route func(Endpoint) echo.HandlerFunc

func CreateJSONEndpoint[T any](
	method string,
	path string,
	handlerFunc func(*prik.Context, *T) *Result,
) Endpoint {
	e := JSONEndpoint{
		method: method,
		path: path,
		handle: func(ctx *prik.Context, c echo.Context) *Result {
			var data T

			if err := c.Bind(&data); err != nil {
				return &Result{Status: http.StatusBadRequest, Body: err.Error()}
			}

			validate := validator.New(validator.WithRequiredStructEnabled())

			if err := validate.Struct(data); err != nil {
				return &Result{Status: http.StatusBadRequest, Body: err.Error()}
			}

			return handlerFunc(ctx, &data)
		},
	}

	return Endpoint{tag: "json", json: &e}
}

func CreateStreamingEndpoint(
	method string,
	path string,
	handlerFunc func(*prik.Context, *http.Request) *http.Response,
) Endpoint {
	e := StreamingEndpoint{
		method: method,
		path: path,
		handle: func(ctx *prik.Context, c echo.Context) *http.Response {
			return handlerFunc(ctx, c.Request())
		},
	}

	return Endpoint{tag: "streaming", streaming: &e}
}

func CreateEndpoints(context *prik.Context, endpoints []Endpoint, server *echo.Echo) {
	for _, e := range endpoints {
		switch e.tag {
		case "json":
			route := createJSONRoute(context)(e)
			server.Add(e.json.method, e.json.path, route)
		case "streaming":
			route := createStreamingRoute(context)(e)
			server.Add(e.streaming.method, e.streaming.path, route)
		}
	}
}

func createJSONRoute(context *prik.Context) Route {
	return func(e Endpoint) echo.HandlerFunc {
		return func(c echo.Context) error {
			res := e.json.handle(context, c)
			return c.JSON(200, res)
		}
	}
}

func createStreamingRoute(context *prik.Context) Route {
	return func(e Endpoint) echo.HandlerFunc {
		return func(c echo.Context) error {
			res := e.streaming.handle(context, c)
			return c.Stream(200, "application/octet-stream", res.Body)
		}
	}
}
