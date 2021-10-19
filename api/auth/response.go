package auth

type Response map[string] interface{}
type ErrorResponse struct {
	Code string
	Message string
}