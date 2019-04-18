package client

import (
	"github.com/kucjac/jsonapi/query/scope"
	"io"
	"strings"
)

// Delete implements jsonapi repository.Deleter interface
func (r *Repository) Delete(s *scope.Scope) error {
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

	if err := r.do(s, "DELETE", sb.String(), nil, nil, r.deleter(s)); err != nil {
		return err
	}

	return nil
}

func (r *Repository) deleter(s *scope.Scope) func(io.Reader, int) error {
	return func(ro io.Reader, status int) error {

		return nil
	}
}
