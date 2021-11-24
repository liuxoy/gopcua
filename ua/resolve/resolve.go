package resolve

import (
	"sync"

	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/errors"
)

// Lookup returns the id from a list of namespace URLs or an error.
// The values are not parsed and are compared case-sensitive.
func Lookup(s string, ns []string) (uint16, error) {
	for id, nsu := range ns {
		if nsu == s {
			return uint16(id), nil
		}
	}
	return 0, errors.Errorf("unknown namespace url %s", s)
}

// StringArrayResolver provides a static namespace resolver.
//
// The implementation is safe for concurrent access but the
// namespace list must not be changed after initialization.
type StringArrayResolver struct {
	Namespaces []string
}

// Lookup returns the id of the namespace url or an error.
// The values are not parsed and are compared case-sensitive.
func (r *StringArrayResolver) Lookup(s string) (uint16, error) {
	return Lookup(s, r.Namespaces)
}

// ClientResolver implements a dynamic caching namespace resolver
// which uses the namespace array of the server and caches it
// for faster lookup. It is the responsibility of the client
// to refresh the list as required. The lookup is safe for
// concurrent access.
type ClientResolver struct {
	Client *opcua.Client

	mu sync.Mutex
	ns []string
}

// Refresh updates the list of namespaces from the server.
func (r *ClientResolver) Refresh() error {
	ns, err := r.Client.NamespaceArray()
	if err != nil {
		return err
	}

	r.mu.Lock()
	r.ns = ns
	r.mu.Unlock()

	return nil
}

// Lookup returns the id of the namespace url or an error.
// The values are not parsed and are compared case-sensitive.
func (r *ClientResolver) Lookup(s string) (uint16, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return Lookup(s, r.ns)
}
