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
	ret := config.DataMap{}
	if meta := getMetadata(s); meta != nil {
		ret["ecsTaskMetadata"] = meta
	}
	return ret
}

func getMetadata(s *setting) interface{} {
	if s.endpoint == "" {
		return nil
	}

	b, cancel := s.Start(context.Background())
	defer cancel()
	for i := 1; backoff.Continue(b); i++ {
		metadata, err := getMetadataOnce(s)
		if err == nil {
			return metadata
		}
		s.Logf("[%d]: unable to get ecs metadata response: %v", i, err)
	}
	s.Logf("max retries count reached")
	return nil
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
	var endpoint string
	if e := os.Getenv("ECS_CONTAINER_METADATA_URI"); e != "" {
		endpoint = e + "/task"
	}
	if e := os.Getenv("ECS_CONTAINER_METADATA_URI_V4"); e != "" {
		endpoint = e + "/task"
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

func WithEnableV2() Option {
	return func(s *setting) {
		if s.endpoint == "" {
			s.endpoint = "http://169.254.170.2/v2/metadata" //for v2
		}
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
