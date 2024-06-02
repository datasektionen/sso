package export

import "github.com/a-h/templ"

type Service interface {
	LoginForm() templ.Component
}
