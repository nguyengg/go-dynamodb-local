# Test against DynamoDB local with Testcontainers

Get with:
```shell
go get github.com/nguyengg/go-aws-commons/go-dynamodb-local
```

Usage:
```go
package app

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	local "github.com/nguyengg/go-aws-commons/go-dynamodb-local"
	"github.com/stretchr/testify/require"
)

func TestMyApp(t *testing.T) {
	client := local.Default(t)

	_, err := client.ListTables(t.Context(), &dynamodb.ListTablesInput{})
	require.NoError(t, err)

	// GetItem is a utility method to help do a quick DynamoDB GetItem.
	var avM map[string]types.AttributeValue
	avM = local.GetItem(t, client, "table-name", "pk-name", "pk-value")
	avM = local.GetItem(t, client, "another-table", "pk-name", 4, "sk-name", 7)
	avM = local.GetItem(t, client, "table-with-B-pk", "pk-name", &types.AttributeValueMemberB{})
}

```
