package local

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/require"
)

// GetItem makes a GetItem request and fails the test if the call fails.
//
// The hash key name and value pair must be given. The sort key name and value pair is optional; if given, len(a) must
// be exactly 2.
func GetItem(t *testing.T, client *dynamodb.Client, tableName string, hashKeyName string, hashKeyValue any, a ...any) map[string]types.AttributeValue {
	var (
		input = &dynamodb.GetItemInput{TableName: &tableName, Key: map[string]types.AttributeValue{}}
		err   error
	)

	if input.Key[hashKeyName], err = attributevalue.Marshal(hashKeyValue); err != nil {
		require.NoError(t, err, "marshal hash key error")
	}

	switch n := len(a); n {
	case 0:
	case 2:
		input.Key[a[0].(string)], err = attributevalue.Marshal(a[1])
		require.NoError(t, err, "marshal sort key error")
	default:
		t.Errorf("GetItem called with invalid number of arguments for sort key; expected 0 or 2, got %d", n)
	}

	output, err := client.GetItem(t.Context(), input)
	if err != nil {
		require.NoError(t, err, "get item error")
	}

	return output.Item
}
