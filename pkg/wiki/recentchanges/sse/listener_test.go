package sse_test

import (
	"testing"

	"github.com/leebradley/wikiedit-monitor-fast/pkg/wiki/recentchanges"
	wikisse "github.com/leebradley/wikiedit-monitor-fast/pkg/wiki/recentchanges/sse"
	"github.com/r3labs/sse"

	"github.com/sirupsen/logrus/hooks/test"
)

type inListener struct {
	data string
	err  error
	lo   recentchanges.ListenOptions
}

type wantListener struct {
	rc  wikisse.RecentChange
	err error
}

var parserTests = []struct {
	name string
	in   inListener
	want wantListener
}{
	{
		name: "basic",
		in: inListener{
			data: `{"wiki":"enwiki"}`,
			lo: recentchanges.ListenOptions{
				Hidebots: true,
				Wikis:    []string{"en"},
			},
		},
		want: wantListener{
			rc: wikisse.RecentChange{
				Wiki: "enwiki",
			},
			err: nil,
		},
	},
}

type listenInput struct {
	rc  wikisse.RecentChange
	err error
}

func TestListener(t *testing.T) {
	for _, tt := range parserTests {
		t.Run(tt.name, func(t *testing.T) {
			logger, _ := test.NewNullLogger()
			client := NewFakeSSEClient(tt.in.data, tt.in.err)
			listener := wikisse.NewListener(client, logger)

			in := make(chan listenInput)
			listener.Listen(tt.in.lo, func(rc wikisse.RecentChange, err error) {
				in <- listenInput{
					rc:  rc,
					err: err,
				}
			})
			received := <-in

			if received.rc != tt.want.rc {
				t.Errorf("got %+v, want %+v", received.rc, tt.want.rc)
			}

			if received.err != tt.want.err {
				t.Errorf("got %+v, want %+v", received.err, tt.want.err)
			}
		})
	}
}

type FakeSSEClient struct {
	data string
	err  error
	url  string
}

func NewFakeSSEClient(data string, err error) *FakeSSEClient {
	return &FakeSSEClient{
		data: data,
		err:  err,
	}
}

func (client *FakeSSEClient) Subscribe(url string, handler func(msg *sse.Event)) error {
	client.url = url

	if client.data != "" {
		handler(&sse.Event{
			Data: []byte(client.data),
		})
	}

	return client.err
}
