package diffs

import (
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"
)

type Compare struct {
	FromID    int    `json:"fromid"`
	FromRevID int    `json:"fromrevid"`
	FromNS    int    `json:"fromns"`
	FromTitle string `json:"fromtitle"`
	ToID      int    `json:"toid"`
	ToRevID   int    `json:"torevid"`
	ToNS      int    `json:"tons"`
	ToTitle   string `json:"totitle"`
	Body      string `json:"*"`
}

type CompareResult struct {
	Compare Compare `json:"compare"`
}

type DiffParser struct {
	logger *logrus.Logger
}

func NewDiffParser(logger *logrus.Logger) DiffParser {
	return DiffParser{logger: logger}
}

func (r DiffParser) Parse(input []byte) (Compare, error) {
	result := CompareResult{}
	err := json.Unmarshal(input, &result)
	if err != nil {
		data := string(input[:])
		err := fmt.Errorf("There was an error decoding: %s", data)
		return Compare{}, err
	}
	return result.Compare, nil
}
