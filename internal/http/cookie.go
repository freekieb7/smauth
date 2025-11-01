package http

import "time"

type SameSite uint8

const (
	SameSiteDefaultMode SameSite = iota
	SameSiteLaxMode
	SameSiteStrictMode
	SameSiteNoneMode
)

type Cookie struct {
	Name     string
	Value    string
	Path     string
	Domain   string
	Expires  time.Time
	MaxAge   int
	Secure   bool
	HttpOnly bool
	SameSite SameSite
}
