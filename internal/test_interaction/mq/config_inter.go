package mq

const (
	okState      = "OK"
	noRouteState = "NO_ROUTE"
	failState    = "CONSUMER_FAIL"
	connErrState = "CONN_ERROR"
	timeoutState = "TIMEOUT"
)

const (
	emailRegisterKey = "email.register"
	emailNoRouteKey  = "email.xxx.yyy"
	emailFailKey     = "email.fail"
)

var messCases = []messCase{
	{okState, emailRegisterKey},
	{noRouteState, emailNoRouteKey},
	{failState, emailFailKey},
	{timeoutState, emailRegisterKey},
	{connErrState, emailRegisterKey},
}

type messCase struct {
	name string
	key  string
}
