package serviceWrappers

// HTTPRequester interface defines the operations supported by a component able to make HTTP requests
type HTTPRequester interface {
	DoRequest(method string, url string, apiKey string, result any) error
	IsInterfaceNil() bool
}
