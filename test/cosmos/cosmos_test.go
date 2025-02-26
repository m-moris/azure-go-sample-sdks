package cosmos

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos"
	"github.com/stretchr/testify/assert"
)

const (
	// Cosmos エミュレーターの接続文字列（Go SDK はHTTP 接続が許可されているの）
	connectionString = "AccountEndpoint=http://localhost:8081/;AccountKey=C2y6yDjf5/R+ob0N8A7Cgv30VRDJIWEHLM+4QDU5DE2nQ9nDuVTqobD4b8mGGyPMbIZnqyMsEcaGQy67XIw/Jw=="
)

func NewClient() *azcosmos.Client {

	client, err := azcosmos.NewClientFromConnectionString(connectionString, nil)
	if err != nil {
		panic(err)
	}

	return client
}

func TestCosmos(t *testing.T) {

	client := NewClient()
	ctx := context.Background()

	// データベースを作成する
	th := azcosmos.NewManualThroughputProperties(400)
	_, err := client.CreateDatabase(ctx,
		azcosmos.DatabaseProperties{
			ID: "testdb",
		},
		&azcosmos.CreateDatabaseOptions{
			ThroughputProperties: &th,
		})

	if err != nil && !isConflict(err) {
		assert.NoError(t, err, "failed to create database")
	}

	// コンテナを作成する
	database, err := client.NewDatabase("testdb")
	assert.NoError(t, err)
	_, err = database.CreateContainer(ctx,
		azcosmos.ContainerProperties{
			ID: "testcontainer",
			PartitionKeyDefinition: azcosmos.PartitionKeyDefinition{
				Paths: []string{"/pk"},
			},
		}, &azcosmos.CreateContainerOptions{})
	if err != nil && !isConflict(err) {
		assert.NoError(t, err, "failed to create container")
	}

	// アイテムを作成する

	item := Item{
		Id:           "1",
		PartitionKey: "1",
		Category:     "category1",
		Name:         "name1",
		Quantity:     10,
		Price:        100.0,
		Clearance:    true,
	}
	container, err := database.NewContainer("testcontainer")
	assert.NoError(t, err)

	// アイテムの作成
	bin, err := json.Marshal(item)
	assert.NoError(t, err)
	partitionKey := azcosmos.NewPartitionKeyString("1")
	_, err = container.CreateItem(ctx, partitionKey, bin, nil)
	assert.NoError(t, err, "failed to create item")

	// アイテムの取得
	resp, err := container.ReadItem(ctx, partitionKey, "1", nil)
	assert.NoError(t, err, "failed to read item")

	var readItem Item
	err = json.Unmarshal(resp.Value, &readItem)
	assert.NoError(t, err, "failed to unmarshal item")
	t.Logf("item: %+v", readItem)

	// クエリ
	pager := container.NewQueryItemsPager("SELECT * FROM c", partitionKey, &azcosmos.QueryOptions{})
	for pager.More() {
		response, err := pager.NextPage(ctx)
		assert.NoError(t, err, "failed to query items")

		for _, v := range response.Items {

			var item Item
			err := json.Unmarshal(v, &item)
			assert.NoError(t, err, "failed to unmarshal item")
			t.Logf("item: %+v", item)

			// アイテムの削除
			_, err = container.DeleteItem(ctx, partitionKey, item.Id, &azcosmos.ItemOptions{})
			assert.NoError(t, err, "failed to delete item")

		}
	}
}

// ゴミ掃除用
func TestDelete(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	// データベースを削除する
	dataqbase, _ := client.NewDatabase("testdb")
	container, _ := dataqbase.NewContainer("testcontainer")
	container.DeleteItem(ctx, azcosmos.NewPartitionKeyString("1"), "1", nil)
}

func isConflict(err error) bool {

	var responseErr *azcore.ResponseError
	errors.As(err, &responseErr)
	return responseErr.StatusCode == 409
}
