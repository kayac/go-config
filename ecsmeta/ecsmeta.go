package ecsmeta

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/kayac/go-config"
	"github.com/lestrrat-go/backoff"
	"github.com/pkg/errors"
)

func New(opts ...Option) config.DataMap {
	s := newSetting()
	for _, opt := range opts {
		opt(s)
	}
	return config.DataMap{
		"ecsTaskMetadata": getMetadata(s),
	}
}

func getMetadata(s *setting) interface{} {
	b, cancel := s.Start(context.Background())
	defer cancel()
	for i := 1; backoff.Continue(b); i++ {
		metadata, err := getMetadataOnce(s)
		if err == nil {
			return metadata
		}
		s.Logf("[%d]: unable to get ecs metadata response: %v", i, err)
	}
	panic("max retries count reached")
}

func getMetadataOnce(s *setting) (interface{}, error) {
	resp, err := s.client.Get(s.endpoint)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get response")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("incorrect status code %d", resp.StatusCode)
	}
	defer resp.Body.Close()
	var metadata interface{}
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&metadata); err != nil {
		return nil, errors.Wrap(err, "unable to decode response body")
	}
	return metadata, nil
}

type Option func(*setting)

type setting struct {
	client   *http.Client
	endpoint string
	policy   backoff.Policy
	logger   *log.Logger
}

func newSetting() *setting {
	endpoint := os.Getenv("ECS_CONTAINER_METADATA_URI") + "/task" //for v3
	if endpoint == "/task" {
		endpoint = "http://169.254.170.2/v2/metadata" //for v2
	}
	return &setting{
		endpoint: endpoint,
		client:   http.DefaultClient,
	}
}

func (s *setting) Logf(format string, args ...interface{}) {
	if s.logger == nil {
		return
	}
	s.logger.Printf(format, args...)
}

func (s *setting) Start(ctx context.Context) (backoff.Backoff, backoff.CancelFunc) {
	if s.policy != nil {
		return s.policy.Start(ctx)
	}
	defaultPolicy := backoff.NewExponential(
		backoff.WithInterval(500*time.Millisecond),
		backoff.WithJitterFactor(0.5),
		backoff.WithMaxRetries(5),
	)
	return defaultPolicy.Start(ctx)
}

func WithEndpoint(endpoint string) Option {
	return func(s *setting) {
		s.endpoint = endpoint
	}
}

func WithHTTPClient(client *http.Client) Option {
	return func(s *setting) {
		s.client = client
	}
}

func WithRetryPolicy(policy backoff.Policy) Option {
	return func(s *setting) {
		s.policy = policy
	}
}

func WithLogger(l *log.Logger) Option {
	return func(s *setting) {
		s.logger = l
	}
}
