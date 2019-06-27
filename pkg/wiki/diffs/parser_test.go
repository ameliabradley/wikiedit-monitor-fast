package diffs_test

import (
	"testing"

	"github.com/leebradley/wikiedit-monitor-fast/pkg/wiki/diffs"
	"github.com/sirupsen/logrus/hooks/test"
)

type inParser struct {
	data string
}

type wantParser struct {
	compare diffs.Compare
	err     error
}

var parserTests = []struct {
	name string
	in   inParser
	want wantParser
}{
	{
		name: "basic",
		in: inParser{
			data: `{"compare":{"fromid":31530695,"fromrevid":903665607,"fromns":2,"fromtitle":"User:DeltaQuad/UAA/Time","toid":31530695,"torevid":903668373,"tons":2,"totitle":"User:DeltaQuad/UAA/Time","*":"test"}}`,
		},
		want: wantParser{
			compare: diffs.Compare{
				FromID:    31530695,
				FromRevID: 903665607,
				FromNS:    2,
				FromTitle: "User:DeltaQuad/UAA/Time",
				ToID:      31530695,
				ToRevID:   903668373,
				ToNS:      2,
				ToTitle:   "User:DeltaQuad/UAA/Time",
				Body:      "test",
			},
			err: nil,
		},
	},
}

func TestParser(t *testing.T) {
	for _, tt := range parserTests {
		t.Run(tt.name, func(t *testing.T) {
			logger, _ := test.NewNullLogger()
			parser := diffs.NewDiffParser(logger)
			compare, err := parser.Parse([]byte(tt.in.data))
			if err != tt.want.err {
				t.Errorf("got %q, want %q", err, tt.want.err)
			}

			if compare != tt.want.compare {
				t.Errorf("got %v, want %v", compare, tt.want.compare)
			}
		})
	}
}
