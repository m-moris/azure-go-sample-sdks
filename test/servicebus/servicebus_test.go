package servicebus

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
	"github.com/stretchr/testify/assert"
)

const (
	connectionString = "Endpoint=sb://localhost;SharedAccessKeyName=RootManageSharedAccessKey;SharedAccessKey=SAS_KEY_VALUE;UseDevelopmentEmulator=true;"
)

func NewClient() *azservicebus.Client {

	client, err := azservicebus.NewClientFromConnectionString(connectionString, &azservicebus.ClientOptions{})
	if err != nil {
		panic(err)
	}
	return client
}

func TestServiceBus(t *testing.T) {

	ctx := context.Background()
	client := NewClient()

	sender, err := client.NewSender("queue.1", &azservicebus.NewSenderOptions{})
	assert.NoError(t, err)
	defer sender.Close(ctx)

	for i := range 10 {
		message := fmt.Sprintf("Hello, World! %d: %s", i, time.Now().Format("2006-01-02 15:04:05"))
		err = sender.SendMessage(ctx, &azservicebus.Message{Body: []byte(message)}, nil)
		assert.NoError(t, err)
	}

	receiver, err := client.NewReceiverForQueue("queue.1", &azservicebus.ReceiverOptions{})
	assert.NoError(t, err)
	defer receiver.Close(ctx)

	messages, err := receiver.ReceiveMessages(ctx, 20, nil)
	assert.NoError(t, err)
	for _, message := range messages {
		body := message.Body
		fmt.Printf("%s\n", string(body))

		err = receiver.CompleteMessage(context.TODO(), message, nil)
		if err != nil {
			panic(err)
		}
	}

}
