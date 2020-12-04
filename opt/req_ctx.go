package opt

import (
	"context"
	"io"
	"net/http"
)

type ResponseHandleFunc func(statusCode int, stream io.Reader) (interface{}, error)
type ValidateHandleFunc func(interface{}) error
type StatusHandleFunc func(statusCode int) error

type RequestContext struct {
	Ctx                    context.Context           // default: context.TODO()
	ExecuteClient          *http.Client              // default: http.DefaultClient
	Url                    string                    // request url
	Method                 string                    // http method,default: http.MethodGet
	Headers                http.Header               // request header
	ResponseStatusHandler  StatusHandleFunc          // default: always return nil
	LazyRequestBodyHandler func() (io.Reader, error) // lazy eval body provider
	ResponseHandler        ResponseHandleFunc        // default: nil, return interface{} -> []byte
	AfterRequestHandlers   []func()                  // after request
	ValidateFunc           ValidateHandler           // validator func
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
	// migrate headers
	for k, v := range this.Headers {
		vLen := len(v)
		if vLen == 1 {
			request.Header.Set(k, v[0])
		} else if vLen > 1 {
			for _, vIt := range v {
				request.Header.Add(k, vIt)
			}
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
