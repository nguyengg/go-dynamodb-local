package local

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/require"
)

func TestGetItem(t *testing.T) {
	client := Default(t)

	_, err := client.CreateTable(t.Context(), &dynamodb.CreateTableInput{
		TableName: aws.String("test"),
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("pk"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("sk"),
				AttributeType: types.ScalarAttributeTypeN,
			},
		},
		BillingMode: types.BillingModePayPerRequest,
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("pk"),
				KeyType:       types.KeyTypeHash,
			},
			{
				AttributeName: aws.String("sk"),
				KeyType:       types.KeyTypeRange,
			},
		},
	})
	require.NoError(t, err)

	// put then get to verify item's attributes.
	item := map[string]types.AttributeValue{
		"pk":   &types.AttributeValueMemberS{Value: "hello"},
		"sk":   &types.AttributeValueMemberN{Value: "7"},
		"data": &types.AttributeValueMemberS{Value: "world"},
	}
	_, err = client.PutItem(t.Context(), &dynamodb.PutItemInput{TableName: aws.String("test"), Item: item})
	require.NoError(t, err)
	require.Equal(
		t,
		item,
		GetItem(t, client, "test", "pk", "hello", "sk", 7))

	// get a non-existent item to verify empty response.
	require.Empty(t, GetItem(t, client, "test", "pk", "hello", "sk", 8))
}
