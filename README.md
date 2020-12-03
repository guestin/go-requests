# go-requests 😉

> inspired by python-requests

![gopher](https://golang.google.cn/lib/godoc/images/footer-gopher.jpg)

## Get Start 🚀

### Install 📦

```bash
go get -u "https://github.com/guestin/go-requests" latest
```

### Usage

we have some bean definations...

```go
// git tag
type GitTag struct {
	Ref    string    `json:"ref" validate:"required"`
	NodeId string    `json:"node_id" validate:"required"`
	TagUrl string    `json:"url validate:"url"`
	Object SHAObject `json:"object" validate:"required"`
}

// SHA object
type SHAObject struct {
	Sha  string `json:"sha" validate:"required"`
	Type string `json:"type" validate:"required"`
	Url  string `json:"url" validate:"url"`
}

```

### simple get

```go
func TestGet(t *testing.T) {
	data, err := Get(context.TODO(),
		"https://api.github.com",
		opt.BuildUrl(murl.WithPath("repos/guestin/mob/git/refs/tags")),
		opt.ExpectStatusCode(http.StatusOK))
	assert.NoError(t, err)
	t.Log(data)
}
```
### retrieve and validate response

work with [go-playground/validator.v10](https://pkg.go.dev/gopkg.in/go-playground/validator.v10) pretty good

```go
func TestValidateResponse(t *testing.T) {
	validator, err := mvalidate.NewValidator("zh")
	assert.NoError(t, err)
	data, err := Get(context.TODO(),
		"https://api.github.com",
		opt.BuildUrl(murl.WithPath("repos/guestin/mob/git/refs/tags")),
		opt.ValidateVar(validator, `required,dive,required`),
		opt.ExpectStatusCode(http.StatusOK),
		opt.BindJSON(&[]GitTag{}))
	if err != nil {
		t.Log(err)
		t.FailNow()
		return
	}
	versionData := data.(*[]GitTag)
	t.Logf("%v\n", versionData)
}

```

### scoped request

```go

// define scope: RemoteGitAPI
var RemoteGitAPI = NewScope(
	opt.Url("https://api.github.com"))

func TestGetScope(t *testing.T) {
	requestScope := RemoteGitAPI.With(
		opt.HttpMethod(http.MethodGet),
		opt.BuildUrl(murl.WithPath("repos/guestin/mob/git/refs/tags")),
		opt.BindJSON(&[]GitTag{}))
	data, err := requestScope.Execute()
	assert.NoError(t, err)
	t.Log(data)
}


```

### others...

`go-requests` have a lot of options, just try it yourself! 😎

## License

MIT

