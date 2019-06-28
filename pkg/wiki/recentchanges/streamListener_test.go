package recentchanges_test

import (
	"testing"

	"github.com/leebradley/wikiedit-monitor-fast/pkg/wiki/recentchanges"
	"github.com/r3labs/sse"

	"github.com/sirupsen/logrus/hooks/test"
)

type inListener struct {
	data string
	err  error
	lo   recentchanges.ListenOptions
}

type wantListener struct {
	rc recentchanges.RecentChange
}

var parserTests = []struct {
	name string
	in   inListener
	want wantListener
}{
	{
		name: "basic",
		in: inListener{
			data: `{"compare":{"fromid":31530695,"fromrevid":903665607,"fromns":2,"fromtitle":"User:DeltaQuad/UAA/Time","toid":31530695,"torevid":903668373,"tons":2,"totitle":"User:DeltaQuad/UAA/Time","*":"test"}}`,
		},
		want: wantListener{
			rc: recentchanges.RecentChange{},
		},
	},
}

func TestListener(t *testing.T) {
	for _, tt := range parserTests {
		t.Run(tt.name, func(t *testing.T) {
			logger, _ := test.NewNullLogger()
			client := NewFakeSSEClient(tt.in.data, tt.in.err)
			listener := recentchanges.NewStreamListener(client, logger)
			listener.Listen(tt.in.lo, func(rc recentchanges.RecentChange) {
				if rc != tt.want.rc {
					t.Errorf("got %+v, want %+v", rc, tt.want.rc)
				}
			})
		})
	}
}

type FakeSSEClient struct {
	data string
	err  error
}

func NewFakeSSEClient(data string, err error) *FakeSSEClient {
	return &FakeSSEClient{
		data: data,
		err:  err,
	}
}

func (client *FakeSSEClient) Subscribe(stream string, handler func(msg *sse.Event)) error {
	if client.data != "" {
		handler(&sse.Event{
			Data: []byte(client.data),
		})
	}

	return client.err
}
