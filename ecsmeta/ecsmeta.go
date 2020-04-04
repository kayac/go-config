package ecsmeta

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/kayac/go-config"
	"github.com/pkg/errors"
)

type setting struct {
	client                 *http.Client
	endpoint               string
	maxRetries             int
	durationBetweenRetries time.Duration
	logger                 *log.Logger
}

func (s *setting) Logf(format string, args ...interface{}) {
	if s.logger == nil {
		return
	}
	s.logger.Printf(format, args...)
}

type Option func(*setting)

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
	for i := 0; i < s.maxRetries+1; i++ {
		metadata, err := getMetadataOnce(s)
		if err == nil {
			return metadata
		}
		s.Logf("[%d/%d]: unable to get ecs metadata response: %v", i+1, s.maxRetries+1, err)
		time.Sleep(s.durationBetweenRetries)
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

func newSetting() *setting {
	endpoint := os.Getenv("ECS_CONTAINER_METADATA_URI") + "/task" //for v3
	if endpoint == "/task" {
		endpoint = "http://169.254.170.2/v2/metadata" //for v2
	}
	return &setting{
		endpoint:               endpoint,
		client:                 http.DefaultClient,
		maxRetries:             5,
		durationBetweenRetries: time.Second,
	}
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

func WithMaxRetries(n int) Option {
	return func(s *setting) {
		s.maxRetries = n
	}
}

func WithDurationBetweenRetries(d time.Duration) Option {
	return func(s *setting) {
		s.durationBetweenRetries = d
	}
}

func WithLogger(l *log.Logger) Option {
	return func(s *setting) {
		s.logger = l
	}
}
