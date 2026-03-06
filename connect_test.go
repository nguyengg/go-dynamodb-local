package local

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/stretchr/testify/require"
)

func TestConnect(t *testing.T) {
	client := Connect(t, WithStubCredentialsProvider())

	_, err := client.ListTables(t.Context(), &dynamodb.ListTablesInput{})
	require.NoError(t, err)
}
