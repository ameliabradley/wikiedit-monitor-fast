package diffs_test

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/leebradley/wikiedit-monitor-fast/pkg/wiki/diffs"
	"github.com/sirupsen/logrus/hooks/test"
)

type inFetcher struct {
	revision int
	body     string
}

type wantFetcher struct {
	url  string
	body string
	err  error
}

var fetcherTests = []struct {
	name string
	in   inFetcher
	want wantFetcher
}{
	{
		name: "basic",
		in: inFetcher{
			revision: 100,
			body:     "foo",
		},
		want: wantFetcher{
			url:  "https://en.wikipedia.org/w/api.php?action=compare&format=json&fromrev=100&torelative=prev",
			body: "foo",
			err:  nil,
		},
	},
}

// RoundTripFunc
type RoundTripFunc func(req *http.Request) *http.Response

// RoundTrip .
func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

// NewTestClient returns *http.Client with Transport replaced to avoid making real calls
func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: RoundTripFunc(fn),
	}
}

func TestDiffFetcher(t *testing.T) {
	for _, tt := range fetcherTests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewTestClient(func(req *http.Request) *http.Response {
				url := req.URL.String()
				if url != tt.want.url {
					t.Errorf("got %q, want %q", url, tt.want.url)
				}

				return &http.Response{
					StatusCode: 200,
					// Send response to be tested
					Body: ioutil.NopCloser(bytes.NewBufferString(tt.in.body)),
					// Must be set to non-nil value or it panics
					Header: make(http.Header),
				}
			})

			logger, _ := test.NewNullLogger()
			fetcher := diffs.NewDiffFetcher(logger, *client)
			body, err := fetcher.Fetch(tt.in.revision)
			if err != tt.want.err {
				t.Errorf("got %q, want %q", err, tt.want.err)
			}

			if string(body) != tt.want.body {
				t.Errorf("got %q, want %q", string(body), tt.want.body)
			}
		})
	}
}
