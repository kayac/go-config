package ecsmeta_test

import (
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"github.com/kayac/go-config"
	"github.com/kayac/go-config/ecsmeta"
)

func TestNew(t *testing.T) {
	server := newTestServer()
	defer server.Close()
	os.Setenv("ECS_CONTAINER_METADATA_URI", server.URL+"/v3")
	testNew(t, "v3", []string{"amazon/amazon-ecs-pause:0.1.0", "nrdlngr/nginx-curl"})
	os.Setenv("ECS_CONTAINER_METADATA_URI", server.URL+"/v4")
	testNew(t, "v4", []string{"mreferre/eksutils"})

}

func testNew(t *testing.T, label string, expected []string) {
	t.Run(label, func(t *testing.T) {
		loader := config.New()
		loader.Data(ecsmeta.New(
			ecsmeta.WithLogger(log.New(os.Stderr, "", log.LstdFlags)),
		))

		src := []byte(`
Images:
  {{ range  $c := .ecsTaskMetadata.Containers }}
  - {{ $c.Image }}
  {{ end }}
`)
		c := make(map[string][]string)
		if err := loader.LoadWithEnvBytes(&c, src); err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(c["Images"], expected) {
			t.Errorf("failed to get container images: %#v", c)
		}
	})
}

func newTestServer() *httptest.Server {
	busyCount := 0
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if busyCount < 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			busyCount++
			return
		}
		switch req.URL.Path {
		case "/v4/task":
			w.WriteHeader(http.StatusOK)
			io.WriteString(w, sampleResponseV4)
		case "/v3/task":
			w.WriteHeader(http.StatusOK)
			io.WriteString(w, sampleResponse)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

//this is sample response, from https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-metadata-endpoint-v3.html
const sampleResponse = `
{
  "Cluster": "default",
  "TaskARN": "arn:aws:ecs:us-east-2:012345678910:task/9781c248-0edd-4cdb-9a93-f63cb662a5d3",
  "Family": "nginx",
  "Revision": "5",
  "DesiredStatus": "RUNNING",
  "KnownStatus": "RUNNING",
  "Containers": [
    {
      "DockerId": "731a0d6a3b4210e2448339bc7015aaa79bfe4fa256384f4102db86ef94cbbc4c",
      "Name": "~internal~ecs~pause",
      "DockerName": "ecs-nginx-5-internalecspause-acc699c0cbf2d6d11700",
      "Image": "amazon/amazon-ecs-pause:0.1.0",
      "ImageID": "",
      "Labels": {
        "com.amazonaws.ecs.cluster": "default",
        "com.amazonaws.ecs.container-name": "~internal~ecs~pause",
        "com.amazonaws.ecs.task-arn": "arn:aws:ecs:us-east-2:012345678910:task/9781c248-0edd-4cdb-9a93-f63cb662a5d3",
        "com.amazonaws.ecs.task-definition-family": "nginx",
        "com.amazonaws.ecs.task-definition-version": "5"
      },
      "DesiredStatus": "RESOURCES_PROVISIONED",
      "KnownStatus": "RESOURCES_PROVISIONED",
      "Limits": {
        "CPU": 0,
        "Memory": 0
      },
      "CreatedAt": "2018-02-01T20:55:08.366329616Z",
      "StartedAt": "2018-02-01T20:55:09.058354915Z",
      "Type": "CNI_PAUSE",
      "Networks": [
        {
          "NetworkMode": "awsvpc",
          "IPv4Addresses": [
            "10.0.2.106"
          ]
        }
      ]
    },
    {
      "DockerId": "43481a6ce4842eec8fe72fc28500c6b52edcc0917f105b83379f88cac1ff3946",
      "Name": "nginx-curl",
      "DockerName": "ecs-nginx-5-nginx-curl-ccccb9f49db0dfe0d901",
      "Image": "nrdlngr/nginx-curl",
      "ImageID": "sha256:2e00ae64383cfc865ba0a2ba37f61b50a120d2d9378559dcd458dc0de47bc165",
      "Labels": {
        "com.amazonaws.ecs.cluster": "default",
        "com.amazonaws.ecs.container-name": "nginx-curl",
        "com.amazonaws.ecs.task-arn": "arn:aws:ecs:us-east-2:012345678910:task/9781c248-0edd-4cdb-9a93-f63cb662a5d3",
        "com.amazonaws.ecs.task-definition-family": "nginx",
        "com.amazonaws.ecs.task-definition-version": "5"
      },
      "DesiredStatus": "RUNNING",
      "KnownStatus": "RUNNING",
      "Limits": {
        "CPU": 512,
        "Memory": 512
      },
      "CreatedAt": "2018-02-01T20:55:10.554941919Z",
      "StartedAt": "2018-02-01T20:55:11.064236631Z",
      "Type": "NORMAL",
      "Networks": [
        {
          "NetworkMode": "awsvpc",
          "IPv4Addresses": [
            "10.0.2.106"
          ]
        }
      ]
    }
  ],
  "PullStartedAt": "2018-02-01T20:55:09.372495529Z",
  "PullStoppedAt": "2018-02-01T20:55:10.552018345Z",
  "AvailabilityZone": "us-east-2b"
}
`

//this is sample v4 response, https://docs.aws.amazon.com/ja_jp/AmazonECS/latest/developerguide/task-metadata-endpoint-v4.html
const sampleResponseV4 = `{
    "Cluster": "arn:aws:ecs:us-west-2:&ExampleAWSAccountNo1;:cluster/default",
    "TaskARN": "arn:aws:ecs:us-west-2:&ExampleAWSAccountNo1;:task/default/febee046097849aba589d4435207c04a",
    "Family": "query-metadata",
    "Revision": "7",
    "DesiredStatus": "RUNNING",
    "KnownStatus": "RUNNING",
    "Limits": {
        "CPU": 0.25,
        "Memory": 512
    },
    "PullStartedAt": "2020-03-26T22:25:40.420726088Z",
    "PullStoppedAt": "2020-03-26T22:26:22.235177616Z",
    "AvailabilityZone": "us-west-2c",
    "Containers": [
        {
            "DockerId": "febee046097849aba589d4435207c04aquery-metadata",
            "Name": "query-metadata",
            "DockerName": "query-metadata",
            "Image": "mreferre/eksutils",
            "ImageID": "sha256:1b146e73f801617610dcb00441c6423e7c85a7583dd4a65ed1be03cb0e123311",
            "Labels": {
                "com.amazonaws.ecs.cluster": "arn:aws:ecs:us-west-2:&ExampleAWSAccountNo1;:cluster/default",
                "com.amazonaws.ecs.container-name": "query-metadata",
                "com.amazonaws.ecs.task-arn": "arn:aws:ecs:us-west-2:&ExampleAWSAccountNo1;:task/default/febee046097849aba589d4435207c04a",
                "com.amazonaws.ecs.task-definition-family": "query-metadata",
                "com.amazonaws.ecs.task-definition-version": "7"
            },
            "DesiredStatus": "RUNNING",
            "KnownStatus": "RUNNING",
            "Limits": {
                "CPU": 2
            },
            "CreatedAt": "2020-03-26T22:26:24.534553758Z",
            "StartedAt": "2020-03-26T22:26:24.534553758Z",
            "Type": "NORMAL",
            "Networks": [
                {
                    "NetworkMode": "awsvpc",
                    "IPv4Addresses": [
                        "10.0.0.108"
                    ],
                    "AttachmentIndex": 0,
                    "IPv4SubnetCIDRBlock": "10.0.0.0/24",
                    "MACAddress": "0a:62:17:7a:36:68",
                    "DomainNameServers": [
                        "10.0.0.2"
                    ],
                    "DomainNameSearchList": [
                        "us-west-2.compute.internal"
                    ],
                    "PrivateDNSName": "ip-10-0-0-108.us-west-2.compute.internal",
                    "SubnetGatewayIpv4Address": ""
                }
            ]
        }
    ]
}
`
