package opt

import (
	"context"
	"github.com/guestin/mob/merrors"
	"io"
	"net/http"
)

type ChainResponseHandleFunc func(statusCode int, stream io.Reader, previousValue interface{}) (interface{}, error)
type ChainedResponseHandler func(next ChainResponseHandleFunc) ChainResponseHandleFunc

type ValidateHandleFunc func(interface{}) error
type CustomRequestHandleFunc func(req *http.Request) error

type RequestContext struct {
	Ctx                       context.Context             // default: context.TODO()
	ExecuteClient             *http.Client                // default: http.DefaultClient
	Url                       string                      // request url
	Method                    string                      // http method,default: http.MethodGet
	LazyRequestBodyHandler    func() (io.Reader, error)   // lazy eval body provider
	responseHandlers          [3][]ChainedResponseHandler // chained handlers
	CustomHttpRequestHandlers []CustomRequestHandleFunc   // after request build
	DeferHandlers             []func()                    // after request executed
}

type InstallPosition int

const (
	HEAD InstallPosition = 0
	PROC InstallPosition = 1
	TAIL InstallPosition = 2
)

func (this *RequestContext) BuildResponseHandler() ChainResponseHandleFunc {
	var chainedHandlers []ChainedResponseHandler
	for _, it := range this.responseHandlers {
		if len(it) == 0 {
			continue
		}
		chainedHandlers = append(chainedHandlers, it...)
	}
	var rspHandler ChainResponseHandleFunc
	for i := len(chainedHandlers) - 1; i >= 0; i-- {
		it := chainedHandlers[i]
		rspHandler = it(rspHandler)
	}
	return rspHandler
}

func (this *RequestContext) InstallResponseHandler(f ChainResponseHandleFunc, pos InstallPosition) {
	responseHandler := func(next ChainResponseHandleFunc) ChainResponseHandleFunc {
		return func(statusCode int, stream io.Reader, previousValue interface{}) (interface{}, error) {
			v, err := f(statusCode, stream, previousValue)
			if err != nil {
				return nil, err
			}
			if next == nil {
				return v, nil
			}
			return next(statusCode, stream, v)
		}
	}
	merrors.Assert(pos < 3, "bad position")
	target := &this.responseHandlers[pos]
	*target = append(*target, responseHandler)
}

func (this *RequestContext) BuildRequest() (*http.Request, error) {
	var err error
	bodyReader := io.Reader(nil)
	if this.LazyRequestBodyHandler != nil {
		bodyReader, err = this.LazyRequestBodyHandler()
		if err != nil {
			return nil, err
		}
	}
	// create newly http.Request
	request, err := http.NewRequestWithContext(
		this.Ctx,
		this.Method,
		this.Url,
		bodyReader)
	if err != nil {
		return nil, err
	}
	// invoke custom request handlers
	for _, customReqHandler := range this.CustomHttpRequestHandlers {
		if err = customReqHandler(request); err != nil {
			return nil, err
		}
	}
	return request, nil
}

type Option func(*RequestContext) error

func (this Option) Concat(cbs ...Option) Option {
	return func(options *RequestContext) error {
		var err error
		cbs = append([]Option{this}, cbs...)
		for _, op := range cbs {
			if err = op(options); err != nil {
				return err
			}
		}
		return nil
	}
}
