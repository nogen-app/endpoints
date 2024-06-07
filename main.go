package endpoints

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/nogen-app/prik"
)

type Result struct {
	Status int `json:"status"`
	Body string `json:"body"`
}

type Endpoint struct {
	Method string
	Path string
	Handle func(*prik.Context, *http.Request) Result
}

type Route func(Endpoint) echo.HandlerFunc

type ContextReq func(*http.Request) (*prik.Context, prik.DisposeFn)
type HandlerFunc func(*prik.Context, *http.Request) Result

func CreateEndpoint(
	method string,
	path string,
	handlerFunc HandlerFunc,
) Endpoint {
	return Endpoint{
		Method: method,
		Path:   path,
		Handle: func(ctx *prik.Context, req *http.Request) Result {
			return handlerFunc(ctx, req)
		},
	}
}

func CreateEndpoints(context ContextReq, endpoints []Endpoint, server *echo.Echo) {
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

func createRoute(context ContextReq) Route {
	return func(e Endpoint) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			ctx, dispose := context(req)
			defer dispose()
			res := e.Handle(ctx, req)
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

// IMPLEMENTATON TESTS

// type TestUser struct {
// 	Username string `json:"username" validate:"required"`
// }

// type Config struct {
// 	databaseUrl string
// 	usersUrl    string
// }

// func LoadConfig() *Config {
// 	return &Config{
// 		databaseUrl: "pgsql://test:test",
// 		usersUrl:    "http://nogen.blocks/users",
// 	}
// }

// type Services struct {
// 	database string
// 	usersApi string
// }

// type UserAPI struct {
// 	GetUser    func(string) string
// 	CreateUser func(TestUser) (TestUser, error)
// }

// func DBFactory(config *Config) prik.Factory {
// 	return func() (interface{}, prik.DisposeFn) {
// 		dispose := func() {}
// 		// TODO: access stuff from config to construct db
// 		return "blockdb", dispose
// 	}
// }

// func UserAPIFactory(db prik.Factory) prik.Factory {
// 	return func() (interface{}, prik.DisposeFn) {
// 		dispose := func() {}
// 		return UserAPI{
// 				GetUser:    func(name string) string { return name },
// 				CreateUser: func(user TestUser) (TestUser, error) { return user, nil },
// 			},
// 			dispose
// 	}
// }

// func base(config *Config) func() *prik.Context {
// 	dbFactory := prik.Shared(DBFactory(config))

// 	return func() *prik.Context {
// 		factories := prik.Factories{
// 			"db":      dbFactory,
// 			"userApi": prik.Shared(UserAPIFactory(dbFactory)),
// 		}

// 		return prik.CreateContext(factories)
// 	}
// }

// func fromRequest(fn func() *prik.Context) func(*http.Request) (*prik.Context, prik.DisposeFn) {
// 	return func(req *http.Request) (*prik.Context, prik.DisposeFn) {
// 		// TODO: this is not right .. figure out what to do with the http.Request and also the dispose
// 		ctx := fn()
// 		return ctx, ctx.Dispose
// 	}
// }

// func testCreateUserHandler(ctx *prik.Context, req *http.Request) Result {
// 	userApi, err := ctx.Resolve("userApi")

// 	if err != nil {
// 		return Result{
// 				Status: 500,
// 				Body:   "Failed to resolve UserAPI from context",
// 			}
// 	}

// 	payload, err := DecodeJSONBody[TestUser](req)

// 	if err != nil {
// 		return Result{Status: 500, Body: err.Error()}
// 	}

// 	api := userApi.(UserAPI)

// 	res := api.GetUser(payload.Username)

// 	return Result{Status: 200, Body: res}
// }

// func getEndpoints() []Endpoint {
// 	createUserEndpoint := CreateEndpoint(
// 		"POST",
// 		"/users",
// 		testCreateUserHandler,
// 	)

// 	endpoints := make([]Endpoint, 0)
// 	endpoints = append(endpoints, createUserEndpoint)
// 	return endpoints
// }

// func createApp(config *Config) *echo.Echo {
// 	// construct the context
// 	ctx := fromRequest(base(config))
// 	server := echo.New()
// 	endpoints := getEndpoints()

// 	CreateEndpoints(ctx, endpoints, server)

// 	return server
// }

// func main() {
// 	// construct a config (load from env etc)
// 	config := LoadConfig()

// 	// Create the app (http server)
// 	app := createApp(config)

// 	// Start server and listen
// 	app.Logger.Fatal(app.Start(":1323"))
// }
