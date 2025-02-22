package endpoints

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"reflect"
	"strconv"

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
			b := new(CustomDataBinder)

			var data T

			if err := b.CustomDataBinder(&data, c); err != nil {
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

func ApplyEndpoints(context *prik.Context, endpoints []Endpoint, server *echo.Echo, m ...echo.MiddlewareFunc) {
	for _, e := range endpoints {
		switch e.tag {
		case "json":
			route := createJSONRoute(context)(e)
			server.Add(e.json.method, e.json.path, route, m...)
		case "streaming":
			route := createStreamingRoute(context)(e)
			server.Add(e.streaming.method, e.streaming.path, route, m...)
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

type CustomDataBinder struct {
	echo.DefaultBinder
}

func (b *CustomDataBinder) CustomDataBinder(i interface{}, c echo.Context) error {
	if err := b.Bind(i, c); err != nil {
		return err
	}

	if err := b.BindHeaders(c, i); err != nil {
		return err
	}

	if err := bindFiles(i, c); err != nil {
		return err
	}

	return nil
}

func bindFiles(i interface{}, c echo.Context) error {
	val := reflect.ValueOf(i)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	typ := val.Type()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		if field.Type == reflect.TypeOf((*multipart.FileHeader)(nil)) {
			formTag := field.Tag.Get("form"); if formTag == "" {
				continue
			}
			maxSizeTag := field.Tag.Get("maxSize")

			file, err := c.FormFile(formTag); if err != nil {
				if err == http.ErrMissingFile {
					continue
				}
				return err
			}

			if maxSizeTag != "" {
				maxSize, err := strconv.ParseInt(maxSizeTag, 10, 64); if err != nil {
					return fmt.Errorf("maxSize tag must be an integer")
				}

				if file.Size > maxSize {
					return fmt.Errorf("file size exceeds the maximum size of %d bytes", maxSize)
				}
			}

			val.Field(i).Set(reflect.ValueOf(file))
		}
	}

	return nil
}
