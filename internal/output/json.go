package output

import (
	"encoding/json"
	"io"

	"github.com/barthollomew/why-is-this-slow/internal/model"
)

type GenericRecord struct {
	Version  string         `json:"version"`
	Run      interface{}    `json:"run"`
	Analysis model.Analysis `json:"analysis"`
}

func WriteJSON(out io.Writer, run interface{}, analysis model.Analysis) error {
	rec := GenericRecord{
		Version:  "1.0",
		Run:      run,
		Analysis: analysis,
	}
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	return enc.Encode(rec)
}
