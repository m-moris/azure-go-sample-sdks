package table

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/data/aztables"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

const (
	connectionString = "DefaultEndpointsProtocol=http;AccountName=devstoreaccount1;AccountKey=Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==;BlobEndpoint=http://127.0.0.1:40000/devstoreaccount1;QueueEndpoint=http://127.0.0.1:40001/devstoreaccount1;TableEndpoint=http://127.0.0.1:40002/devstoreaccount1;"
	uri              = "http://127.0.0.1:40002/devstoreaccount1"
	containerName    = "testcontainer"
)

func CreateServiceClient() *aztables.ServiceClient {

	// キュー用のサービスクラアントを作成する
	serviceClient, err := aztables.NewServiceClientFromConnectionString(connectionString, nil)
	if err != nil {
		panic(err)
	}
	return serviceClient
}

func TestTable(t *testing.T) {

	svc := CreateServiceClient()
	ctx := context.Background()
	name := "testtable"

	// テーブルの作成、既に存在する場合のエラーは無視する
	_, err := svc.CreateTable(ctx, name, nil)
	if err != nil && !isTableAlreadyExistsError(err) {
		assert.Error(t, err)
	}

	// テーブルクライアントの作成
	client := svc.NewClient(name)

	// 生エンティティを作る
	myEntity := aztables.EDMEntity{
		Entity: aztables.Entity{
			PartitionKey: uuid.NewString(),
			RowKey:       "RowKey",
		},
		Properties: map[string]any{
			"Stock":                15,
			"Price":                9.99,
			"Comments":             "great product",
			"OnSale":               true,
			"ReducedPrice":         7.99,
			"PurchaseDate":         aztables.EDMDateTime(time.Date(2021, time.August, 21, 1, 1, 0, 0, time.UTC)),
			"BinaryRepresentation": aztables.EDMBinary([]byte("Bytesliceinfo")),
		},
	}

	// マーシャリングして、エンティティを追加する
	marshalled, err := json.Marshal(myEntity)
	assert.NoError(t, err)
	_, err = client.AddEntity(context.TODO(), marshalled, nil)
	assert.NoError(t, err)

	// 構造体を定義してエンティティを追加する
	myEntity2 := MyEntity{
		PartitionKey: uuid.NewString(),
		RowKey:       "RowKey1",
		Stock:        15,
		Price:        9.99,
		Comments:     "great product",
		OnSale:       true,
		ReducedPrice: 7.99,
		PurchaseDate: time.Date(2021, time.August, 21, 1, 1, 0, 0, time.UTC),
		BinaryRep:    []byte("Bytesliceinfo"),
	}

	marshalled, err = json.Marshal(myEntity2)
	assert.NoError(t, err)

	_, err = client.AddEntity(ctx, marshalled, nil)
	assert.NoError(t, err)

	// クエリ パラメータを使ってエンティティを取得する
	filter := fmt.Sprintf("RowKey eq '%s'", "RowKey")
	options := &aztables.ListEntitiesOptions{
		Filter: &filter,
	}
	pager := client.NewListEntitiesPager(options)
	for pager.More() {
		response, err := pager.NextPage(ctx)
		assert.NoError(t, err)

		for _, entityBytes := range response.Entities {
			var entity MyEntity
			err := json.Unmarshal(entityBytes, &entity)
			assert.NoError(t, err)

			t.Logf("Found entity: %v", entity)

			// エンティティの削除
			client.DeleteEntity(ctx, entity.PartitionKey, entity.RowKey, nil)
		}
	}

	// テーブルの削除
	//svc.DeleteTable(ctx, name, nil)
}

// エラーがテーブルが既に存在するエラーかどうかを判定する
func isTableAlreadyExistsError(err error) bool {
	var respErr *azcore.ResponseError
	if errors.As(err, &respErr) {
		return respErr.ErrorCode == string(aztables.TableAlreadyExists)
	}
	return false
}
