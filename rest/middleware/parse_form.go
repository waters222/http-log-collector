package middleware

import (
	"github.com/weishi258/http-log-collector/log"
	"github.com/weishi258/http-log-collector/rest/constants"
	"github.com/weishi258/http-log-collector/rest/error_code"
	"go.uber.org/zap"
	"net/http"
)

func ParseForm(handler http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			log.GetLogger().Error("parse form failed", zap.String("error", err.Error()))
			w.Header().Set(constants.ContentType, constants.ContentTypeDefault)
			w.WriteHeader(http.StatusInternalServerError)
			if _, err = w.Write(error_code.NewResponseError(error_code.InternalError).MarshalJsonByte()); err != nil {
				log.GetLogger().Error("marshal error response to json failed", zap.String("error", err.Error()))
			}
		} else {
			handler(w, r)
		}

	})

}
