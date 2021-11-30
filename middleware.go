package hevent

import (
	"fmt"

	"github.com/kamva/hexa"
	"github.com/kamva/tracer"
)

type MiddlewareFunc func(handler EventHandler) EventHandler

// WithMiddlewares adds middlewares to the handler too.
func WithMiddlewares(h EventHandler, middlewares ...MiddlewareFunc) EventHandler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

// RecoverMiddleware is a event handler middleware which recover panic error.
func RecoverMiddleware(h EventHandler) EventHandler {
	return func(hc HandlerContext, c hexa.Context, message Message, err error) (errResult error) {
		defer func() {
			if r := recover(); r != nil {
				err, ok := r.(error)
				if !ok {
					err = fmt.Errorf("recovered unknown panic: %s", r)
				}

				errResult = tracer.Trace(err)
			}
		}()

		return tracer.Trace(h(hc, c, message, err))
	}
}
