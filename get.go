package client

import (
	"errors"
	"github.com/kucjac/jsonapi/encoding/jsonapi"
	"github.com/kucjac/jsonapi/query/scope"
	"io"
	"strings"
)

// ErrNoPrimaryFilters is an error thrown when the scope doesn't contain primary filters
var ErrNoPrimaryFilters = errors.New("Scope has no primary filters")

// Get implements the jsonapi repository.Getter interface
func (r *Repository) Get(s *scope.Scope) error {

	primID, err := r.getPrimID(s)
	if err != nil {
		return err
	}

	sb := &strings.Builder{}
	sb.WriteString("/")
	sb.WriteString(s.Struct().SchemaName())
	sb.WriteString("/")
	sb.WriteString(s.Struct().Collection())
	sb.WriteString("/")
	sb.WriteString(primID)

	if err := r.do(s, "GET", sb.String(), nil, nil, r.getter(s)); err != nil {
		return err
	}

	return nil
}

// getter is the reader function for the provided scope
func (r *Repository) getter(s *scope.Scope) func(io.Reader, int) error {
	return func(reader io.Reader, status int) error {
		return jsonapi.UnmarshalRegisteredC(s.Controller(), reader, s.Value)
	}
}
