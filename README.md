# Go SDK Samples for Azure

[日本語(Japanese)](README_ja.md)

For those who feel that it is difficult to create and test resources on Azure due to the lack of Go samples in Japanese, we provide samples that allow you to easily check the operation and debug the SDK in a local environment.

## Target SDKs

| Service Name  |     |
| ------------- | --- |
| Storage Blob  | :o: |
| Storage Queue | :o: |
| Storage Table | :o: |
| Cosmos        | :o: |
| Service Bus   | :o: |


## Execution
Start the storage emulator with Docker Compose and test it.
You can check the operation while debugging with a test runner in an IDE such as Visual Studio Code.

```bash
$ docker compose up
$ go test -v -count=1 ./...
```

## References

- [Azure/azure-sdk-for-go: This repository is for active development of the Azure SDK for Go. For consumers of the SDK we recommend visiting our public developer docs at:](https://github.com/Azure/azure-sdk-for-go)


## Notes

- There are old Go libraries, so be careful not to confuse them.
- Error handling is tested with `assert.Error(t, err)`, but please perform appropriate error handling in actual applications.