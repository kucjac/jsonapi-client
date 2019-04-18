package client

import (
	"compress/flate"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/neuronlabs/neuron/controller"
	"github.com/neuronlabs/neuron/encoding/jsonapi"
	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/query/scope"
	"gopkg.in/go-playground/validator.v9"
	"io"
	"net/http"
	"net/url"
	"strings"
)

var validate = validator.New()

// Repository is the http.Client that implements neuron.Repository
// allowing to query different jsonapi server.
type Repository struct {
	Config *Config

	// Client is the http.Client
	Client *http.Client

	c *controller.Controller
}

// New creates the repository for the provided config
func New(c *controller.Controller, cfg *Config) (*Repository, error) {
	r := &Repository{
		c:      c,
		Config: cfg,
		Client: &http.Client{},
	}

	// validate the input
	if err := r.validate(); err != nil {
		return nil, err
	}

	// check the connection
	if err := r.checkConnection(); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *Repository) checkConnection() error {
	req, err := http.NewRequest("GET", r.getURL("health"), nil)
	if err != nil {
		return err
	}

	resp, err := r.Client.Do(req)
	if err != nil {
		return err
	}

	h := &health{}
	d := json.NewDecoder(resp.Body)
	if err = d.Decode(h); err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Health Response Status not OK: %d", resp.StatusCode)
	}

	if h.Status != "pass" {
		return fmt.Errorf("Invalid health status: %s", h.Status)
	}

	return nil
}

func (r *Repository) do(
	s *scope.Scope,
	method, path string,
	query url.Values,
	data io.Reader, valueReader func(io.Reader, int) error,
) error {
	path = r.getURL(path)
	req, err := http.NewRequest(method, path, data)
	if err != nil {
		log.Error("[CLIENT] %s - %s creating new request failed: %v", method, path, err.Error())
		return err
	}

	if err = r.formatRequest(req); err != nil {
		return err
	}

	req = req.WithContext(s.Context())

	// set the raw query
	if query != nil {
		req.URL.RawQuery = query.Encode()
	}

	// Client do
	resp, err := r.Client.Do(req)
	if err != nil {
		return err
	}

	var reader io.Reader
	if resp.Uncompressed {
		switch enc := resp.Header.Get("Content-Encoding"); enc {
		case "gzip":
			reader, err = gzip.NewReader(resp.Body)
			if err != nil {
				return err
			}
		case "deflate":
			d := flate.NewReader(resp.Body)
			defer d.Close()
			reader = d
		case "":
			reader = resp.Body
			defer resp.Body.Close()
		default:
			log.Errorf("Unsupported Content-Encoding: %s", enc)
			return ErrUnsupportedContentEncoding
		}
	} else {
		reader = resp.Body
	}

	// check if there is an error
	if resp.StatusCode >= 400 {
		if resp.ContentLength != 0 {
			d := json.NewDecoder(reader)
			errs := &jsonapi.ErrorsPayload{}
			if err = d.Decode(&errs); err != nil {
				return err
			}

			return errors.MultipleErrors(errs.Errors)
		}
		switch resp.StatusCode {
		case 400:
			return errors.ErrBadRequest.Copy()
		case 403:
			return errors.ErrEndpointForbidden.Copy()
		case 404:
			return errors.ErrResourceNotFound.Copy()
		case 405:
			return errors.ErrMethodNotAllowed.Copy().WithDetail(fmt.Sprintf("The method: '%s'.", method))
		case 406:
			return errors.ErrNotAcceptable.Copy()
		case 409:
			return errors.ErrResourceAlreadyExists.Copy()
		case 500:
			return errors.ErrInternalError.Copy().WithDetail("Server encountered undefined internal error")
		case 503:
			return errors.ErrServerBusy1.Copy()
		default:
			return (&errors.ApiError{Code: "UNDEFINED", Title: "Undefined Error"}).WithStatus(resp.StatusCode)
		}
	}

	if err = valueReader(reader, resp.StatusCode); err != nil {
		return err
	}

	return nil
}

func (r *Repository) formatRequest(req *http.Request) error {
	// Add all headers
	req.Header.Add("User-Agent", "Golang JSONAPI Client - github.com/neuronlabs/neuron-client")
	req.Header.Add("Accept", "application/vnd.api+json")
	req.Header.Add("Accept-Encoding", "gzip,deflate")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Cache-Control", "max-age=0, no-cache")
	req.Header.Add("Pragma", "no-cache")

	return nil
}

// getURL gets the url values for given schema
func (r *Repository) getURL(path string) string {

	sb := &strings.Builder{}

	sb.WriteString("http")

	if r.Config.HTTPS {
		sb.WriteString("s")
	}

	sb.WriteString(fmt.Sprintf("://%s:%d/v%d", r.Config.Hostname, r.Config.Port, r.Config.APIVersion))
	if r.Config.PathBase != "" {
		sb.WriteString("/")
		sb.WriteString(r.Config.PathBase)
	}

	sb.WriteString(path)

	return sb.String()
}

func (r *Repository) getPrimID(s *scope.Scope) (string, error) {
	prims := s.PrimaryFilters()
	if len(prims) < 1 {
		return "", ErrNoPrimaryFilters
	}

	q := prims[0].FormatQuery()

	var primID string
	for _, v := range q {

		primID = v[0]
		if comma := strings.IndexRune(primID, ','); comma != -1 {
			primID = primID[:comma]
		}
	}

	if primID == "" {
		log.Errorf("[CLIENT] no primaryID value found within filters. PrimaryFilters: %v", q)
		return "", ErrNoPrimaryFilters
	}

	return primID, nil
}

func (r *Repository) validate() error {
	if err := validate.Struct(r.Config); err != nil {
		return err
	}
	return nil
}

type health struct {
	Status string `json:"status"`
}
