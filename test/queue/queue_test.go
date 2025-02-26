package queue

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azqueue"
)

const (
	connectionString = "DefaultEndpointsProtocol=http;AccountName=devstoreaccount1;AccountKey=Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==;BlobEndpoint=http://127.0.0.1:40000/devstoreaccount1;QueueEndpoint=http://127.0.0.1:40001/devstoreaccount1;TableEndpoint=http://127.0.0.1:40002/devstoreaccount1;"
	uri              = "http://127.0.0.1:400011/devstoreaccount1"
	containerName    = "testcontainer"
)

func CreateServiceClient() *azqueue.ServiceClient {

	if connectionString != "" {

		// キュー用のサービスクラアントを作成する
		serviceClient, err := azqueue.NewServiceClientFromConnectionString(connectionString, nil)
		if err != nil {
			panic(err)
		}
		return serviceClient
	} else {
		creds, err := azidentity.NewDefaultAzureCredential(nil)
		if err != nil {
			panic(err)
		}

		serviceClient, err := azqueue.NewServiceClient(uri, creds, nil)
		if err != nil {
			panic(err)
		}
		return serviceClient
	}
}

func TestQueueOperation(t *testing.T) {

	svc := CreateServiceClient()
	ctx := context.Background()
	qname := "testqueue"

	// キューの作成
	_, err := svc.CreateQueue(ctx, qname, &azqueue.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create queue: %v", err)
	}

	client := svc.NewQueueClient(qname)

	errCh := make(chan error)

	// Go ルーチンで メッセージループを回す
	go func() {
		for {
			// メッセージを受信
			messages, err := client.DequeueMessages(ctx, &azqueue.DequeueMessagesOptions{
				NumberOfMessages:  to.Ptr(int32(2)),  // 一度に取得するメッセージ数
				VisibilityTimeout: to.Ptr(int32(60)), // メッセージの非表示期間
			})
			if err != nil {
				errCh <- fmt.Errorf("Failed to dequeue messages: %v", err)
				return
			}
			t.Logf("Dequeued count %d", len(messages.Messages))

			// 受信したメッセージを処理しつつ削除
			for _, message := range messages.Messages {
				t.Logf("Dequeued message: %s", *message.MessageText)
				_, err := client.DeleteMessage(ctx, *message.MessageID, *message.PopReceipt, nil)
				if err != nil {
					errCh <- fmt.Errorf("Failed to delete message: %v", err)
					return
				}
			}

			time.Sleep(1 * time.Second) // 1秒待機してから次のメッセージを受信
		}
	}()
	for i := range 10 {
		res, err := client.EnqueueMessage(ctx, fmt.Sprintf("Hello, World  %d: %s", i, time.Now().Format(time.RFC3339)), nil)
		if err != nil {
			t.Fatalf("Failed to enqueue message: %v", err)
		}
		t.Logf("Enqueued message: %s", *res.RequestID)
	}

	select {
	case err := <-errCh:
		t.Fatal(err)
	case <-time.After(15 * time.Second):
		// テストが終了するまで若干待機する
	}
}
