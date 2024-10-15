package testscommon

import "errors"

// HostsFinderStub -
type HostsFinderStub struct {
	FindHostCalled func(urlValues map[string][]string) (string, error)
}

// FindHost -
func (stub *HostsFinderStub) FindHost(urlValues map[string][]string) (string, error) {
	if stub.FindHostCalled != nil {
		return stub.FindHostCalled(urlValues)
	}

	return "", errors.New("not implemented")
}

// IsInterfaceNil -
func (stub *HostsFinderStub) IsInterfaceNil() bool {
	return stub == nil
}
