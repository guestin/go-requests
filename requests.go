package requests

import (
	"context"
	"github.com/guestin/go-requests/opt"
	"github.com/guestin/mob/mio"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"net/http"
)

func Get(ctx context.Context, url string, opts ...opt.Option) (interface{}, error) {
	return easyWrap(ctx, http.MethodGet, url, opts...)
}

func Post(ctx context.Context, url string, opts ...opt.Option) (interface{}, error) {
	return easyWrap(ctx, http.MethodPost, url, opts...)
}

func Put(ctx context.Context, url string, opts ...opt.Option) (interface{}, error) {
	return easyWrap(ctx, http.MethodPut, url, opts...)
}

func Delete(ctx context.Context, url string, opts ...opt.Option) (interface{}, error) {
	return easyWrap(ctx, http.MethodDelete, url, opts...)
}

func Head(ctx context.Context, url string, opts ...opt.Option) (interface{}, error) {
	return easyWrap(ctx, http.MethodHead, url, opts...)
}

func easyWrap(ctx context.Context, method, url string, opts ...opt.Option) (interface{}, error) {
	opts = append([]opt.Option{
		opt.BindContext(ctx),
		opt.HttpMethod(method),
		opt.Url(url),
	}, opts...)
	return SendRequest1(opts)
}

var defaultRequestParams = opt.RequestContext{
	Ctx:           context.TODO(),
	ExecuteClient: http.DefaultClient,
	Method:        http.MethodGet,
	ResponseStatusHandler: func(statusCode int) error {
		return nil
	},
	LazyRequestBodyHandler: nil,
	ResponseHandler: func(_ int, stream io.Reader) (interface{}, error) {
		dataBytes, err := ioutil.ReadAll(stream)
		if err != nil {
			return nil, err
		}
		return string(dataBytes), nil
	},
}

var ErrNoExecuteClient = errors.New("no http client provide")

func SendRequest(opts ...opt.Option) (interface{}, error) {
	return SendRequest1(opts)
}

func SendRequest1(opts []opt.Option) (interface{}, error) {
	var err error
	reqParam := defaultRequestParams
	for _, itOpt := range opts {
		if err = itOpt(&reqParam); err != nil {
			return nil, errors.Wrap(err, "build request failed")
		}
	}
	httpRequest, err := reqParam.BuildRequest()
	if err != nil {
		return nil, err
	}
	if reqParam.ExecuteClient == nil {
		return nil, ErrNoExecuteClient
	}
	defer func() {
		for _, deferFunc := range reqParam.AfterRequestHandlers {
			deferFunc()
		}
	}()
	httpResp, err := reqParam.ExecuteClient.Do(httpRequest)
	if err != nil {
		return nil, err
	}
	defer mio.CloseIgnoreErr(httpResp.Body)
	statusCode := httpResp.StatusCode
	if reqParam.ResponseStatusHandler != nil {
		err = reqParam.ResponseStatusHandler(statusCode)
		if err != nil {
			return nil, err
		}
	}
	outV, err := reqParam.ResponseHandler(statusCode, httpResp.Body)
	if err != nil {
		return nil, err
	}
	if reqParam.ValidateFunc != nil {
		err = reqParam.ValidateFunc(outV)
		if err != nil {
			// validate error
			return nil, err
		}
	}
	return outV, nil
}
