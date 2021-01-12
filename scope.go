package requests

import "github.com/guestin/go-requests/opt"

type Scope []opt.Option

func NewScope(opts ...opt.Option) Scope {
	return opts
}

func (this Scope) With(opts ...opt.Option) Scope {
	newScope := make(Scope, 0, len(this)+len(opts))
	newScope = append(newScope, this...)
	newScope = append(newScope, opts...)
	return newScope
}

func (this Scope) Execute() (interface{}, error) {
	return SendRequest1(this)
}
