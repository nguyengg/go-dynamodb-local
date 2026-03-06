// Package local provides testing functionality using [DynamoDB local as Docker image].
//
// The DynamoDB local image is started via [Testcontainers DynamoDB module]. Docker daemon must have been started prior
// to executing the tests. On Linux, Docker can be installed via system's default package manager. On MacOS and Windows,
// use Docker Desktop if possible. If using Colima on MacOS, be sure to follow [Using Colima with Docker] to set
// DOCKER_HOST and/or TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE accordingly.
//
// [DynamoDB local as Docker image]: https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/DynamoDBLocal.DownloadingAndRunning.html
// [Testcontainers DynamoDB module]: https://golang.testcontainers.org/modules/dynamodb/.
// [Using Colima with Docker]: https://golang.testcontainers.org/system_requirements/using_colima/
package local

import (
	"context"
	"net/url"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/smithy-go/endpoints"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	testcontainersdynamodb "github.com/testcontainers/testcontainers-go/modules/dynamodb"
)

// Options customises Connect.
type Options struct {
	// Skip can be given to skip the test by calling t.Skip.
	Skip func(t *testing.T)

	// opaque settings.

	containerOptions []testcontainers.ContainerCustomizer
	loadOptions      []func(opts *config.LoadOptions) error
	clientOptions    []func(opts *dynamodb.Options)
}

// Connect should be called by all tests that require a DynamoDB local instance.
func Connect(t *testing.T, optFns ...func(opts *Options)) *dynamodb.Client {
	opts := &Options{}
	for _, fn := range optFns {
		fn(opts)
	}

	if opts.Skip != nil {
		opts.Skip(t)
	}

	c, err := testcontainersdynamodb.Run(context.Background(), "amazon/dynamodb-local:3.3.0", opts.containerOptions...)
	require.NoErrorf(t, err, "start dynamodb-local error")

	cfg, err := config.LoadDefaultConfig(t.Context(), opts.loadOptions...)
	require.NoErrorf(t, err, "load default config error")

	endpoint, err := c.ConnectionString(t.Context())
	require.NoErrorf(t, err, "get connection string to dynamodb-local error")

	client := dynamodb.NewFromConfig(cfg, append(opts.clientOptions, dynamodb.WithEndpointResolverV2(dynamodbLocalEndpoint{endpoint}))...)

	t.Cleanup(func() {
		testcontainers.CleanupContainer(t, c)
	})

	return client
}

// Default is a variation of Connect with sensible defaults.
//
// The settings are enabled by Default:
//   - DynamoDB local is started with -inMemory -sharedDb -disableTelemetry per [usage notes]. This is fine since each
//     test should be creating their own connection.
//   - WithStubCredentialsProvider is added.
//
// [usage notes]: https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/DynamoDBLocal.UsageNotes.html
func Default(t *testing.T, optFns ...func(opts *Options)) *dynamodb.Client {
	return Connect(
		t,
		append([]func(opts *Options){
			WithContainerOptions(testcontainers.WithCmdArgs("-inMemory", "-sharedDb", "-disableTelemetry")),
			WithStubCredentialsProvider(),
		}, optFns...)...)
}

// DefaultSkippable is a variation of Default that skips the test if DOCKER_HOST is not given.
func DefaultSkippable(t *testing.T, optFns ...func(opts *Options)) *dynamodb.Client {
	if os.Getenv("DOCKER_HOST") == "" {
		t.Skip("DOCKER_HOST is not defined")
		return nil
	}

	return Default(t, optFns...)
}

// WithContainerOptions adds customisations to the test container such as [testcontainersdynamodb.WithSharedDB].
func WithContainerOptions(optFns ...testcontainers.ContainerCustomizer) func(opts *Options) {
	return func(opts *Options) {
		opts.containerOptions = append(opts.containerOptions, optFns...)
	}
}

// WithLoadOptions adds customisations to the config.LoadDefaultConfig.
func WithLoadOptions(optFns ...func(opts *config.LoadOptions) error) func(opts *Options) {
	return func(opts *Options) {
		opts.loadOptions = append(opts.loadOptions, optFns...)
	}
}

// WithClientOptions adds client options passed to every DynamoDB calls by the client returned by Connect.
func WithClientOptions(optFns ...func(opts *dynamodb.Options)) func(opts *Options) {
	return func(opts *Options) {
		opts.clientOptions = append(opts.clientOptions, optFns...)
	}
}

// WithStubCredentialsProvider modifies the credentials to use stub values since DynamoDB local doesn't care.
func WithStubCredentialsProvider() func(opts *Options) {
	return func(opts *Options) {
		opts.loadOptions = append(opts.loadOptions, func(opts *config.LoadOptions) error {
			opts.Credentials = credentials.NewStaticCredentialsProvider("stub", "stub", "")
			return nil
		})
	}
}

type dynamodbLocalEndpoint struct {
	hostPort string
}

func (e dynamodbLocalEndpoint) ResolveEndpoint(_ context.Context, _ dynamodb.EndpointParameters) (transport.Endpoint, error) {
	return transport.Endpoint{
		URI: url.URL{
			Host:   e.hostPort,
			Scheme: "http",
		},
	}, nil
}
