package http

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Writer http.ResponseWriter
}

func (r *Response) SetStatus(code StatusCode) {
	r.Writer.WriteHeader(int(code))
}

func (r *Response) SetHeader(key, value string) {
	r.Writer.Header().Set(key, value)
}

func (r *Response) SetCookie(cookie Cookie) {
	http.SetCookie(r.Writer, &http.Cookie{
		Name:     cookie.Name,
		Value:    cookie.Value,
		Path:     cookie.Path,
		Domain:   cookie.Domain,
		MaxAge:   cookie.MaxAge,
		Secure:   cookie.Secure,
		HttpOnly: cookie.HttpOnly,
		SameSite: http.SameSite(cookie.SameSite),
		Expires:  cookie.Expires,
	})
}

func (r *Response) Write(data []byte) (int, error) {
	return r.Writer.Write(data)
}

func (r *Response) Send(code StatusCode) error {
	r.SetStatus(code)
	return nil
}

func (r *Response) SendText(code StatusCode, text string) error {
	r.Writer.Header().Set("Content-Type", "text/plain")
	r.SetStatus(code)
	_, err := r.Writer.Write([]byte(text))
	return err
}

func (r *Response) SendJSON(code StatusCode, data any) error {
	r.Writer.Header().Set("Content-Type", "application/json")
	r.SetStatus(code)
	return json.NewEncoder(r.Writer).Encode(data)
}

func (r *Response) SendRedirect(code StatusCode, url string) error {
	r.SetHeader("Location", url)
	r.SetStatus(code)
	return nil
}
