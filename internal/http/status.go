package http

type StatusCode uint16

const (
	StatusContinue           StatusCode = 100
	StatusSwitchingProtocols StatusCode = 101
	StatusProcessing         StatusCode = 102
	StatusEarlyHints         StatusCode = 103

	StatusOK                   StatusCode = 200
	StatusCreated              StatusCode = 201
	StatusAccepted             StatusCode = 202
	StatusNonAuthoritativeInfo StatusCode = 203
	StatusNoContent            StatusCode = 204
	StatusResetContent         StatusCode = 205
	StatusPartialContent       StatusCode = 206
	StatusMultiStatus          StatusCode = 207
	StatusAlreadyReported      StatusCode = 208
	StatusIMUsed               StatusCode = 226

	StatusMultipleChoices   StatusCode = 300
	StatusMovedPermanently  StatusCode = 301
	StatusFound             StatusCode = 302
	StatusSeeOther          StatusCode = 303
	StatusNotModified       StatusCode = 304
	StatusUseProxy          StatusCode = 305
	StatusTemporaryRedirect StatusCode = 307
	StatusPermanentRedirect StatusCode = 308

	StatusBadRequest                  StatusCode = 400
	StatusUnauthorized                StatusCode = 401
	StatusPaymentRequired             StatusCode = 402
	StatusForbidden                   StatusCode = 403
	StatusNotFound                    StatusCode = 404
	StatusMethodNotAllowed            StatusCode = 405
	StatusNotAcceptable               StatusCode = 406
	StatusProxyAuthRequired           StatusCode = 407
	StatusRequestTimeout              StatusCode = 408
	StatusConflict                    StatusCode = 409
	StatusGone                        StatusCode = 410
	StatusLengthRequired              StatusCode = 411
	StatusPreconditionFailed          StatusCode = 412
	StatusRequestEntityTooLarge       StatusCode = 413
	StatusRequestURITooLong           StatusCode = 414
	StatusUnsupportedMediaType        StatusCode = 415
	StatusRangeNotSatisfiable         StatusCode = 416
	StatusExpectationFailed           StatusCode = 417
	StatusTeapot                      StatusCode = 418
	StatusMisdirectedRequest          StatusCode = 421
	StatusUnprocessableEntity         StatusCode = 422
	StatusLocked                      StatusCode = 423
	StatusFailedDependency            StatusCode = 424
	StatusTooEarly                    StatusCode = 425
	StatusUpgradeRequired             StatusCode = 426
	StatusPreconditionRequired        StatusCode = 428
	StatusTooManyRequests             StatusCode = 429
	StatusRequestHeaderFieldsTooLarge StatusCode = 431
	StatusUnavailableForLegalReasons  StatusCode = 451

	StatusInternalServerError           StatusCode = 500
	StatusNotImplemented                StatusCode = 501
	StatusBadGateway                    StatusCode = 502
	StatusServiceUnavailable            StatusCode = 503
	StatusGatewayTimeout                StatusCode = 504
	StatusHTTPVersionNotSupported       StatusCode = 505
	StatusVariantAlsoNegotiates         StatusCode = 506
	StatusInsufficientStorage           StatusCode = 507
	StatusLoopDetected                  StatusCode = 508
	StatusNotExtended                   StatusCode = 510
	StatusNetworkAuthenticationRequired StatusCode = 511
)
