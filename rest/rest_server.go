package rest

import (
	"context"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/weishi258/http-log-collector/log"
	"github.com/weishi258/http-log-collector/rest/constants"
	"github.com/weishi258/http-log-collector/rest/error_code"
	"github.com/weishi258/http-log-collector/rest/middleware"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type RestServer struct {
	localAddr      string
	server         *http.Server
	router         *mux.Router
	allowedHeaders []string
	allowedOrigins []string
	allowedMethods []string
}

type HandlerFunc func(http.ResponseWriter, *http.Request) ([]byte, int, error_code.ResponseError)
type MiddleWare func(handler http.HandlerFunc) http.HandlerFunc

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc HandlerFunc
	MiddleWares []MiddleWare
}

func getDefaultAllowedHeaders() []string {
	return []string{"*"}
}
func getDefaultAllowedOrigins() []string {
	return []string{"*"}
}
func getDefaultAllowedMethods() []string {
	return []string{"GET", "POST", "OPTIONS", "DELETE", "PUT"}
}
func NewRestServer(localAddr string) *RestServer {
	ret := &RestServer{localAddr: localAddr,
		allowedHeaders: getDefaultAllowedHeaders(),
		allowedOrigins: getDefaultAllowedOrigins(),
		allowedMethods: getDefaultAllowedMethods()}
	ret.router = mux.NewRouter().SkipClean(true)

	return ret
}

func (c *RestServer) Start(sigChan chan bool, keepalive bool) error {
	if c.server != nil {
		return errors.New("rest rest already started")
	}
	headersOk := handlers.AllowedHeaders(c.allowedHeaders)
	originsOk := handlers.AllowedOrigins(c.allowedOrigins)
	methodsOk := handlers.AllowedMethods(c.allowedMethods)
	for _, site := range c.allowedOrigins {
		log.GetLogger().Info("CORS allowed origins", zap.String("site", site))
	}
	for _, header := range c.allowedHeaders {
		log.GetLogger().Info("CORS allowed headers", zap.String("header", header))
	}

	for _, method := range c.allowedMethods {
		log.GetLogger().Info("CORS allowed methods", zap.String("method", method))
	}

	c.server = &http.Server{Addr: c.localAddr, Handler: handlers.CORS(headersOk, originsOk, methodsOk)(c.router)}
	c.server.SetKeepAlivesEnabled(keepalive)

	go func() {
		log.GetLogger().Info("proxy rest started", zap.String("addr", c.localAddr))
		if err := c.server.ListenAndServe(); err != nil {
			log.GetLogger().Info("proxy RestServer stopped", zap.String("addr", c.localAddr), zap.String("cause", err.Error()))
			if sigChan != nil {
				sigChan <- true
			}
		}
	}()
	return nil
}

func (c *RestServer) Shutdown() error {
	if c.server == nil {
		return errors.New("proxy rest not started")
	}
	log.GetLogger().Info("proxy RestServer is shutting down", zap.String("addr", c.localAddr))
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	if err := c.server.Shutdown(ctx); err != nil {
		log.GetLogger().Error("proxy RestServer shutdown failed", zap.String("error", err.Error()))
	} else {
		log.GetLogger().Info("proxy RestServer shutdown successful", zap.String("addr", c.localAddr))
	}
	c.server = nil
	return nil
}

func (c *RestServer) GetRouter() *mux.Router {
	return c.router
}

func (c *RestServer) AddRoutes(routes []Route) {
	for _, route := range routes {
		if route.HandlerFunc != nil {
			handlerFunc := BaseWrapper(route.HandlerFunc)
			if route.MiddleWares != nil && len(route.MiddleWares) > 0 {
				for _, middleWare := range route.MiddleWares {
					handlerFunc = middleWare(handlerFunc)
				}
			}
			// add request logger if in debug mode
			if log.IsDebug() {
				handlerFunc = middleware.Logger(handlerFunc)
			}
			c.router.Methods(route.Method).
				Path(route.Pattern).
				Name(route.Name).
				Handler(handlerFunc)

		}
	}
}

func BaseWrapper(handler HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		responseByte, httpCode, errRes := handler(w, r)
		var err error
		if errRes != nil {
			// only set to application/json when handler did not set it
			if w.Header().Get(constants.ContentType) == "" {
				w.Header().Set(constants.ContentType, constants.ContentTypeDefault)
			}
			if httpCode == 0 {
				w.WriteHeader(http.StatusBadRequest)
			} else {
				w.WriteHeader(httpCode)
			}
			_, err = w.Write(errRes.MarshalJsonByte())
		} else {
			// set application/json as default content-type if missing
			if w.Header().Get(constants.ContentType) == "" {
				w.Header().Set(constants.ContentType, constants.ContentTypeDefault)
			}
			if httpCode == 0 {
				w.WriteHeader(http.StatusOK)
			} else {
				w.WriteHeader(httpCode)
			}

			_, err = w.Write(responseByte)
		}
		if err != nil {
			log.GetLogger().Error("<BaseWrapper> write response failed", zap.String("error", err.Error()))
		}
	})
}
