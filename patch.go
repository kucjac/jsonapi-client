package client

import (
	"bytes"
	"github.com/neuronlabs/neuron/encoding/jsonapi"
	"github.com/neuronlabs/neuron/query/scope"
	"io"
	"net/http"
	"strings"
)

// Patch implements the neuron repository.Patcher interface
func (r *Repository) Patch(s *scope.Scope) error {

	primID, err := r.getPrimID(s)
	if err != nil {
		return err
	}

	b := &bytes.Buffer{}
	if err := jsonapi.MarshalC(s.Controller(), b, s.Value); err != nil {
		return err
	}

	sb := &strings.Builder{}
	sb.WriteString("/")
	sb.WriteString(s.Struct().SchemaName())
	sb.WriteString("/")
	sb.WriteString(s.Struct().Collection())
	sb.WriteString("/")
	sb.WriteString(primID)

	// TODO: support filters

	if err := r.do(s, "PATCH", sb.String(), nil, b, r.patcher(s)); err != nil {
		return err
	}

	return nil
}

func (r *Repository) patcher(s *scope.Scope) func(io.Reader, int) error {
	return func(ro io.Reader, status int) error {
		if status == http.StatusNoContent {
			return nil
		}

		return jsonapi.UnmarshalC(s.Controller(), ro, s.Value)
	}
}
