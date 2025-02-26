package blob

import (
	"context"
	"fmt"
	"io"
	"os"

	"strconv"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/bloberror"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/sas"
	"github.com/stretchr/testify/assert"
)

const (
	connectionString = "DefaultEndpointsProtocol=http;AccountName=devstoreaccount1;AccountKey=Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==;BlobEndpoint=http://127.0.0.1:40000/devstoreaccount1;QueueEndpoint=http://127.0.0.1:40001/devstoreaccount1;TableEndpoint=http://127.0.0.1:40002/devstoreaccount1;"
	uri              = "http://127.0.0.1:40000/devstoreaccount1"
	containerName    = "testcontainer"
)

func NewClient() *azblob.Client {

	if connectionString != "" {
		// 接続文字列からクライアントを作成する
		client, err := azblob.NewClientFromConnectionString(connectionString, nil)
		if err != nil {
			panic(err)
		}
		return client
	} else {
		// 本番などは、NewDefaultAzureCredential を使うと良い
		cred, err := azidentity.NewDefaultAzureCredential(nil)
		if err != nil {
			panic(err)
		}

		client, err := azblob.NewClient(uri, cred, nil)
		if err != nil {
			panic(err)
		}
		return client
	}
}

func TestBlob(t *testing.T) {

	ctx := context.Background()
	client := NewClient()

	// コンテナを作成する
	err := CreateContainerIfNotExists(client, ctx, containerName)
	assert.NoError(t, err, "failed to create container")

	// コンテナの一覧を取得する
	cp := client.NewListContainersPager(&azblob.ListContainersOptions{
		MaxResults: to.Ptr(int32(100)),
	})

	for cp.More() {
		page, err := cp.NextPage(ctx)
		assert.NoError(t, err, "failed to list containers")
		for _, container := range page.ContainerItems {
			t.Logf("container name : %s", *container.Name)
		}
	}

	// BLOBをアップロードする。基本的に上書きされる
	for i := 0; i < 10; i++ {
		_, err := client.UploadBuffer(ctx, containerName, ("testblob" + strconv.Itoa(i)),
			fmt.Appendf(nil, "Hello, World! now=%v", time.Now()),
			&azblob.UploadBufferOptions{})
		assert.NoError(t, err, "failed to upload blob")
	}

	// BLOBの一覧を取得する
	bp := client.NewListBlobsFlatPager(containerName, &azblob.ListBlobsFlatOptions{
		MaxResults: to.Ptr(int32(100)),
	})
	for bp.More() {
		page, err := bp.NextPage(ctx)
		assert.NoError(t, err, "failed to list blobs")
		for _, blob := range page.Segment.BlobItems {
			t.Logf("blob name : %s", *blob.Name)
		}
	}

	// 色々な方法でBLOBをダウンロードする
	dw1, err := client.DownloadStream(ctx, containerName, "testblob0", &azblob.DownloadStreamOptions{})
	assert.NoError(t, err, "failed to download blob")
	buf, err := io.ReadAll(dw1.Body)
	assert.NoError(t, err, "failed to read blob content")
	t.Logf("blob content : %s", string(buf))

	// ファイルにダウンロードする
	os.Remove("testblob1")
	f, err := os.Create("testblob1")
	assert.NoError(t, err, "failed to create file")
	defer f.Close()
	_, err = client.DownloadFile(ctx, containerName, "testblob1", f, &azblob.DownloadFileOptions{
		RetryReaderOptionsPerBlock: blob.RetryReaderOptions{
			MaxRetries: 3,
		},
	})
	assert.NoError(t, err, "failed to download blob to file")
	t.Log("downloaded to testblob1")

	// BLOBを削除する
	for i := range 10 {
		_, err := client.DeleteBlob(ctx, containerName, ("testblob" + strconv.Itoa(i)), &azblob.DeleteBlobOptions{})
		assert.NoError(t, err, "failed to delete blob")
	}

	// コンテナを削除する
	_, err = client.DeleteContainer(ctx, containerName, &azblob.DeleteContainerOptions{})
	assert.NoError(t, err, "failed to delete container")
}

// TestMics は、その他の機能を試す
func TestMics(t *testing.T) {

	ctx := context.Background()
	client := NewClient()

	err := CreateContainerIfNotExists(client, ctx, containerName)
	assert.NoError(t, err, "failed to create container")

	// BLOBを TAG/ContentType付きで アップロードする
	st := time.Now().Add(-1 * time.Minute)
	client.UploadBuffer(ctx, containerName, "testblob", fmt.Appendf(nil, "Hello, World!. time=%v", time.Now()), &azblob.UploadBufferOptions{
		Tags: map[string]string{
			"tag1": "value1",
			"tag2": "value2",
			"tag3": "value3",
			"tag4": "value4",
		},

		HTTPHeaders: &blob.HTTPHeaders{
			BlobContentType: to.Ptr("text/plain"),
		},
	})

	// SAS を以下の条件で生成
	// - 有効期限は24時間
	// - 読み取りのみ許可
	testBlob := client.ServiceClient().NewContainerClient(containerName).NewBlobClient("testblob")
	sasURL, err := testBlob.GetSASURL(
		sas.BlobPermissions{
			Read: true,
		},
		time.Now().Add(24*time.Hour),
		&blob.GetSASURLOptions{
			StartTime: &st,
		})
	assert.NoError(t, err, "failed to generate SAS URL")
	t.Logf("SAS URL: %s", sasURL)

	// BLOBのタグを取得する
	tags, err := testBlob.GetTags(ctx, &blob.GetTagsOptions{})
	assert.NoError(t, err, "failed to get blob properties")
	for _, v := range tags.BlobTags.BlobTagSet {

		t.Logf("tag key=%s, value=%s", *v.Key, *v.Value)
	}

	// プロパティを取得する
	props, err := testBlob.GetProperties(ctx, &blob.GetPropertiesOptions{})
	assert.NoError(t, err, "failed to get blob properties")
	t.Logf("content type: %s", *props.ContentType)
	t.Logf("created on: %v", *props.CreationTime)
	t.Logf("last modified: %v", *props.LastModified)

}

// CreateContainerIfNotExists は、コンテナが存在しない場合にコンテナを作成する
// 既に存在するか否かはエラーで判断する
func CreateContainerIfNotExists(client *azblob.Client, ctx context.Context, containerName string) error {

	_, err := client.CreateContainer(ctx, containerName, nil)
	if err != nil {
		if bloberror.HasCode(err, bloberror.ContainerAlreadyExists) {
			return nil
		} else {
			panic(err)
		}
	}
	return nil
}
