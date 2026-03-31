package formatters

import (
	"encoding/json"
	"io"

	"github.com/mattneto928/deadhead/models"
	"github.com/mattneto928/deadhead/search"
)

// ToJSON writes the object out to the stream
func ToJSON(wr io.Writer, req *models.Request, summaries []*search.CitySummary) error {
	enc := json.NewEncoder(wr)
	enc.SetIndent("", "\t")
	return enc.Encode(struct {
		Request *models.Request           `json:"request"`
		Data    []*search.CitySummary `json:"data"`
	}{
		Request: req,
		Data:    summaries,
	})
}
