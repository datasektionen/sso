package export

import (
	"context"

	"github.com/a-h/templ"
)

type Service interface {
	PasskeyLogin() func() templ.Component
	PasskeySettings(ctx context.Context, kthid string) (func() templ.Component, error)
}
