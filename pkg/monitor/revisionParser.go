package monitor

import (
	"encoding/json"
	"fmt"
	"log"
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

type RevisionParser struct {
	logger *log.Logger
}

func NewRevisionParser(logger *log.Logger) RevisionParser {
	return RevisionParser{logger: logger}
}

func (r RevisionParser) Parse(input []byte) (Compare, error) {
	result := CompareResult{}
	err := json.Unmarshal(input, &result)
	if err != nil {
		data := string(input[:])
		err := fmt.Errorf("There was an error decoding: %s\n", data)
		return Compare{}, err
	}
	return result.Compare, nil
}
