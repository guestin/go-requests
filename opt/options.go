package opt

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"github.com/go-playground/validator/v10"
	"github.com/guestin/mob/murl"
	"github.com/guestin/mob/mvalidate"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"net/http"
)

// append header
func AddHeader(key string, values ...string) Option {
	return EditRequest(func(req *http.Request) error {
		for _, it := range values {
			req.Header.Add(key, it)
		}
		return nil
	})
}

// set header
func SetHeader(key string, value string) Option {
	return EditRequest(func(req *http.Request) error {
		req.Header.Set(key, value)
		return nil
	})
}

// assign http method, INTERNAL API WARNING!
func HttpMethod(method string) Option {
	return func(options *RequestContext) error {
		options.Method = method
		return nil
	}
}

// request's context, default: context.TODO()
func BindContext(ctx context.Context) Option {
	return func(options *RequestContext) error {
		options.Ctx = ctx
		return nil
	}
}

// expect http response code,default is 200
func ExpectStatusCode(statusCodes ...int) Option {
	return func(options *RequestContext) error {
		options.ResponseStatusHandler = func(statusCode int) error {
			if len(statusCodes) == 0 {
				return nil
			}
			for _, it := range statusCodes {
				if statusCode == it {
					return nil
				}
			}
			return errors.Errorf("unexpect status code:%d", statusCode)
		}
		return nil
	}
}

// custom build url
func BuildUrl(urlBuilderOptions ...murl.UrlBuildOption) Option {
	return func(options *RequestContext) error {
		curUrl := options.Url
		if len(curUrl) == 0 {
			return errors.New(
				"no base url, you must set url(call Url) before BuildUrl")
		}
		newUrl, err := murl.MakeUrlString(curUrl, urlBuilderOptions...)
		if err != nil {
			return errors.Wrap(err, "BuildUrl")
		}
		options.Url = newUrl
		return nil
	}
}

// request url
func Url(baseUrl string) Option {
	return func(options *RequestContext) error {
		options.Url = baseUrl
		return nil
	}
}

// set request body
func Body(reader io.Reader) Option {
	return func(options *RequestContext) error {
		options.LazyRequestBodyHandler = func() (io.Reader, error) {
			return reader, nil
		}
		return nil
	}
}

// add some defer function after request done
func Defer(cb ...func()) Option {
	return func(options *RequestContext) error {
		options.AfterRequestHandlers =
			append(options.AfterRequestHandlers, cb...)
		return nil
	}
}

// content type
func ContentType(contentType string) Option {
	return SetHeader("Content-Type", contentType)
}

// body with json contents
func BodyJSON(v interface{}) Option {
	return CustomBody("application/json", json.Marshal, v)
}

// body with xml contents
func BodyXML(v interface{}) Option {
	return CustomBody("application/xml", xml.Marshal, v)
}

// marshal func
type MarshalFunc func(interface{}) ([]byte, error)

// custom body
func CustomBody(contentType string, marshalFunc MarshalFunc, v interface{}) Option {
	return ContentType(contentType).
		Concat(func(options *RequestContext) error {
			options.LazyRequestBodyHandler = func() (io.Reader, error) {
				dataBytes, err := marshalFunc(v)
				if err != nil {
					return nil, err
				}
				dataBuffer := bytes.NewBuffer(dataBytes)
				return dataBuffer, nil
			}
			return nil
		})
}

// add validator0
func Validator(validateIns *validator.Validate) Option {
	return CustomValidator(validateIns.Struct)
}

// mob.Validator var
func ValidateVar(validIns mvalidate.Validator, validateTag string) Option {
	return CustomValidator(func(i interface{}) error {
		return validIns.Var(i, validateTag)
	})
}

// mob.Validator struct
func ValidateStruct(validIns mvalidate.Validator) Option {
	return CustomValidator(validIns.Struct)
}

// custom validator func
func CustomValidator(validateFunc ValidateHandleFunc) Option {
	return func(options *RequestContext) error {
		options.ValidateFunc = validateFunc
		return nil
	}
}

// drop response body, response value is status_code
func DropResponseBody() Option {
	return func(options *RequestContext) error {
		options.ResponseHandler = func(statusCode int, stream io.Reader) (interface{}, error) {
			return statusCode, nil
		}
		return nil
	}
}

// unmarshal func
type UnmarshalFunc func([]byte, interface{}) error

// response data binder
func DataBind(unmarshal UnmarshalFunc, value interface{}) Option {
	return func(options *RequestContext) error {
		options.ResponseHandler = func(_ int, stream io.Reader) (interface{}, error) {
			dataBytes, err := ioutil.ReadAll(stream)
			if err != nil {
				return nil, errors.Wrap(err, "read response stream failed")
			}
			err = unmarshal(dataBytes, value)
			if err != nil {
				return nil, errors.Wrap(err, "unmarshal response data failed")
			}
			return value, nil
		}
		return nil
	}
}

// bind output json
func BindJSON(value interface{}) Option {
	return DataBind(json.Unmarshal, value)
}

// bind output xml
func BindXML(value interface{}) Option {
	return DataBind(xml.Unmarshal, value)
}

// edit http request
func EditRequest(f CustomRequestHandleFunc) Option {
	return func(requestContext *RequestContext) error {
		requestContext.CustomHttpRequestHandlers =
			append(requestContext.CustomHttpRequestHandlers, f)
		return nil
	}
}
