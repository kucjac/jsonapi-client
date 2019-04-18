package client

import (
	"bytes"
	"github.com/neuronlabs/neuron/encoding/jsonapi"
	"github.com/neuronlabs/neuron/query/scope"
	"io"
	"strings"
)

// Create implements the neuron repository.Creator interface
func (r *Repository) Create(s *scope.Scope) error {
	b := &bytes.Buffer{}

	if err := jsonapi.MarshalC(s.Controller(), b, s.Value); err != nil {
		return err
	}

	sb := &strings.Builder{}
	sb.WriteString("/")
	sb.WriteString(s.Struct().SchemaName())
	sb.WriteString("/")
	sb.WriteString(s.Struct().Collection())

	if err := r.do(s, "POST", sb.String(), nil, b, r.creator(s)); err != nil {
		return err
	}

	return nil
}

func (r *Repository) creator(s *scope.Scope) func(io.Reader, int) error {
	return func(ro io.Reader, status int) error {
		return jsonapi.UnmarshalC(s.Controller(), ro, s.Value)
	}
}
