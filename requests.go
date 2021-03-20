package requests

import (
	"context"
	"github.com/guestin/go-requests/opt"
	"github.com/guestin/mob/mio"
	"github.com/pkg/errors"
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
		for _, deferFunc := range reqParam.DeferHandlers {
			deferFunc()
		}
	}()
	httpResp, err := reqParam.ExecuteClient.Do(httpRequest)
	if err != nil {
		return nil, err
	}
	defer mio.CloseIgnoreErr(httpResp.Body)
	statusCode := httpResp.StatusCode
	rspHandler := reqParam.BuildResponseHandler()
	if rspHandler == nil {
		return statusCode, nil
	}
	return rspHandler(statusCode, httpResp.Body, nil)
}
