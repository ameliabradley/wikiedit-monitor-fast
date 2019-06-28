package wiki

import "github.com/r3labs/sse"

// SSEClient implements the subscribe method for an SSE client
type SSEClient interface {
	Subscribe(url string, handler func(msg *sse.Event)) error
}

type sseClient struct{}

// NewSSEClient creates a new SSEClient
func NewSSEClient() SSEClient {
	return &sseClient{}
}

func (s *sseClient) Subscribe(url string, handler func(msg *sse.Event)) error {
	client := sse.NewClient(url)
	client.Subscribe("messages", handler)
	return nil
}
