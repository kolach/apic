package apic

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io"

	"github.com/pkg/errors"
)

// JSONBody encodes arbitrary interface value as JSON string
func JSONBody(i interface{}) (io.Reader, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	if err := enc.Encode(i); err != nil {
		return nil, errors.Wrapf(err, "failed to encode %+v", i)
	}
	return &buf, nil
}

// XMLBody encodes arbitrary interface value as XML
func XMLBody(i interface{}) (io.Reader, error) {
	var buf bytes.Buffer
	enc := xml.NewEncoder(&buf)
	if err := enc.Encode(i); err != nil {
		return nil, errors.Wrapf(err, "failed to encode %+v", i)
	}
	return &buf, nil
}
