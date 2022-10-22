package auth

type Auth interface {
	Authenticate(request Request) (Response, error)
}

type Request struct {
	Data map[string]interface{}
}

type Response struct {
}
