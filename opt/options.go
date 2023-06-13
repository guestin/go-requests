package opt

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/guestin/mob/mio"
	"github.com/guestin/mob/murl"
	"github.com/guestin/mob/mvalidate"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"os"
)

// AddHeader append header
func AddHeader(key string, values ...string) Option {
	return EditRequest(func(req *http.Request) error {
		for _, it := range values {
			req.Header.Add(key, it)
		}
		return nil
	})
}

// SetHeader set header
func SetHeader(key string, value string) Option {
	return EditRequest(func(req *http.Request) error {
		req.Header.Set(key, value)
		return nil
	})
}

// HttpMethod assign http method, INTERNAL API WARNING!
func HttpMethod(method string) Option {
	return func(options *RequestContext) error {
		options.Method = method
		return nil
	}
}

// BindContext request's context, default: context.TODO()
func BindContext(ctx context.Context) Option {
	return func(options *RequestContext) error {
		options.Ctx = ctx
		return nil
	}
}

// ExpectStatusCode expect http response code,default is 200
func ExpectStatusCode(statusCodes ...int) Option {
	return func(options *RequestContext) error {
		options.InstallResponseHandler(func(statusCode int, stream io.ReadSeeker, previousValue interface{}) (interface{}, error) {
			if len(statusCodes) == 0 {
				return nil, nil
			}
			for _, it := range statusCodes {
				if statusCode == it {
					return nil, nil
				}
			}
			return nil, errors.Errorf("unexpect status code:%d", statusCode)
		}, HEAD)
		return nil
	}
}

// BuildUrl custom build url
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

// Url request url
func Url(baseUrl string) Option {
	return func(options *RequestContext) error {
		options.Url = baseUrl
		return nil
	}
}

// Body set request body
func Body(reader io.Reader) Option {
	return func(options *RequestContext) error {
		options.LazyRequestBodyHandler = func() (io.Reader, error) {
			return reader, nil
		}
		return nil
	}
}

// Defer add some defer function after request done
func Defer(cb ...func()) Option {
	return func(options *RequestContext) error {
		options.DeferHandlers =
			append(options.DeferHandlers, cb...)
		return nil
	}
}

// ContentType content type
func ContentType(contentType string) Option {
	return SetHeader("Content-Type", contentType)
}

// BodyJSON body with json contents
func BodyJSON(v interface{}) Option {
	return CustomBody("application/json", json.Marshal, v)
}

// BodyXML body with xml contents
func BodyXML(v interface{}) Option {
	return CustomBody("application/xml", xml.Marshal, v)
}

// MarshalFunc marshal func
type MarshalFunc func(interface{}) ([]byte, error)

// CustomBody custom body
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

// Validator add validator0
func Validator(validateIns *validator.Validate) Option {
	return CustomValidator(validateIns.Struct)
}

// ValidateVar mob.Validator var
func ValidateVar(validIns mvalidate.Validator, validateTag string) Option {
	return CustomValidator(func(i interface{}) error {
		return validIns.Var(i, validateTag)
	})
}

// ValidateStruct mob.Validator struct
func ValidateStruct(validIns mvalidate.Validator) Option {
	return CustomValidator(validIns.Struct)
}

// CustomValidator custom validator func
func CustomValidator(validateFunc ValidateHandleFunc) Option {
	return func(options *RequestContext) error {
		options.InstallResponseHandler(
			func(statusCode int,
				stream io.ReadSeeker,
				previousValue interface{}) (interface{}, error) {
				if previousValue == nil {
					return nil, nil
				}
				if err := validateFunc(previousValue); err != nil {
					return nil, err
				}
				return previousValue, nil
			}, TAIL)
		return nil
	}

}

// DropResponseBody drop response body, response value is status_code
func DropResponseBody() Option {
	return func(options *RequestContext) error {
		options.InstallResponseHandler(
			func(statusCode int, stream io.ReadSeeker, previousValue interface{}) (interface{}, error) {
				return statusCode, nil
			}, PROC)
		return nil
	}
}

// UnmarshalFunc unmarshal func
type UnmarshalFunc func([]byte, interface{}) error

// DataBind response data binder
func DataBind(unmarshal UnmarshalFunc, value interface{}) Option {
	return func(options *RequestContext) error {
		options.InstallResponseHandler(func(_ int, stream io.ReadSeeker, previousValue interface{}) (interface{}, error) {
			dataBytes, err := io.ReadAll(stream)
			if err != nil {
				return nil, errors.Wrap(err, "read response stream failed")
			}
			err = unmarshal(dataBytes, value)
			if err != nil {
				return nil, errors.Wrap(err, "unmarshal response data failed")
			}
			return value, nil
		}, PROC)
		return nil
	}
}

// BindJSON bind output json
func BindJSON(value interface{}) Option {
	return DataBind(json.Unmarshal, value)
}

// BindXML bind output xml
func BindXML(value interface{}) Option {
	return DataBind(xml.Unmarshal, value)
}

// BindString bind body to string
func BindString(str *string) Option {
	return DataBind(func(data []byte, i interface{}) error {
		if s, ok := i.(*string); ok && s != nil {
			*s = string(data)
		}
		return nil
	}, str)
}

// EditRequest edit http request
func EditRequest(f CustomRequestHandleFunc) Option {
	return func(requestContext *RequestContext) error {
		requestContext.CustomHttpRequestHandlers =
			append(requestContext.CustomHttpRequestHandlers, f)
		return nil
	}
}

// ResponseBodyToFile output: nWrite: int64
func ResponseBodyToFile(fileName string, flag int, perm os.FileMode) Option {
	return func(reqCtx *RequestContext) error {
		reqCtx.InstallResponseHandler(func(statusCode int, stream io.ReadSeeker, _ interface{}) (interface{}, error) {
			output, err := os.OpenFile(fileName, flag, perm)
			if err != nil {
				return nil, err
			}
			defer mio.CloseIgnoreErr(output)
			nWrite, err := io.Copy(output, stream)
			if err != nil {
				return nil, err
			}
			return nWrite, nil
		}, PROC)
		return nil
	}
}

// ResponseBodyDump output: nWrite: int64
func ResponseBodyDump(output io.Writer) Option {
	return func(reqCtx *RequestContext) error {
		reqCtx.InstallResponseHandler(func(statusCode int, stream io.ReadSeeker, objectValue interface{}) (interface{}, error) {
			if _, err := fmt.Fprintf(output, "===\nstatus code:%d\n", statusCode); err != nil {
				return nil, err
			}
			//
			if _, err := io.Copy(output, stream); err != nil {
				return nil, err
			}
			//
			if _, err := fmt.Fprintln(output, "==="); err != nil {
				return nil, err
			}
			return objectValue, nil
		}, PROC)
		return nil
	}
}
