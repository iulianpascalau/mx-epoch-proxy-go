package process

// HostFinder is able to return a valid host based on a search criteria
type HostFinder interface {
	FindHost(urlValues map[string][]string) (string, error)
	IsInterfaceNil() bool
}
