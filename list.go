package client

import (
	"github.com/kucjac/jsonapi/encoding/jsonapi"
	"github.com/kucjac/jsonapi/query/scope"
	"io"
	"strings"
)

// List implements the jsonapi repository.Lister
func (r *Repository) List(s *scope.Scope) error {

	sb := &strings.Builder{}
	sb.WriteString("/")
	sb.WriteString(s.Struct().SchemaName())
	sb.WriteString("/")
	sb.WriteString(s.Struct().Collection())

	if err := r.do(s, "GET", sb.String(), s.FormatQuery(), nil, r.lister(s)); err != nil {
		return err
	}

	return nil
}

func (r *Repository) lister(s *scope.Scope) func(io.Reader, int) error {
	return func(ro io.Reader, status int) error {
		return jsonapi.UnmarshalC(s.Controller(), ro, s.Value)
	}
}
