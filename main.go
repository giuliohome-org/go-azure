package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
)

var (
	subscriptionID     string
	location           = "westeurope"
	resourceGroupName  = "go-azure-sdk"
	storageAccountName = "golangazure"
	containerName      = "golang-container-" + RandStringBytes(4)
)

var (
	storageClientFactory *armstorage.ClientFactory
	blobContainersClient *armstorage.BlobContainersClient
)

func main() {
	fmt.Println("Starting azure golang main.")
	subscriptionID = os.Getenv("AZURE_SUBSCRIPTION_ID")
	if len(subscriptionID) == 0 {
		log.Fatal("AZURE_SUBSCRIPTION_ID is not set.")
	}

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		// handle error
		log.Fatal(err)
	}
	ctx := context.Background()

	storageClientFactory, err = armstorage.NewClientFactory(subscriptionID, cred, nil)
	if err != nil {
		log.Fatal(err)
	}
	blobContainersClient = storageClientFactory.NewBlobContainersClient()

	blobContainer, err := getBlobContainer(ctx)

	var respErr *azcore.ResponseError
	if err != nil {
		if errors.As(err, &respErr) {
			// Handle Error
			if respErr.StatusCode == http.StatusNotFound {
				fmt.Printf("Blob container could not be found, return code: %v\n", respErr.ErrorCode)
				fmt.Println("Creating it now...")
				blobContainer_new, err_new := createBlobContainers(ctx)
				if err_new != nil {
					log.Fatal(err_new)
					return
				}
				log.Println("Created blob container:", *blobContainer_new.ID)
				blobContainer_again, err_again := getBlobContainer(ctx)
				if err_again != nil {
					log.Fatal(err_again)
					return
				}
				log.Println("Double check, blob container ID:", *blobContainer_again.ID)
			} else {
				fmt.Printf("Container Get status code %d error code: %v", respErr.StatusCode, respErr.ErrorCode)
				log.Fatal(respErr)
			}
		} else {
			log.Fatal(err)
		}
	} else {
		id := *blobContainer.ID
		fmt.Printf("Blob container already exists, id: %v", id)
	}
	return
}

const letterBytes = "abcdefghijklmnopqrstuvwxyz"

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func createBlobContainers(ctx context.Context) (*armstorage.BlobContainer, error) {

	blobContainerResp, err := blobContainersClient.Create(
		ctx,
		resourceGroupName,
		storageAccountName,
		containerName,
		armstorage.BlobContainer{
			ContainerProperties: &armstorage.ContainerProperties{

				PublicAccess: to.Ptr(armstorage.PublicAccessNone),
			},
		},
		nil,
	)

	if err != nil {
		return nil, err
	}

	return &blobContainerResp.BlobContainer, nil
}

func getBlobContainer(ctx context.Context) (blobContainer *armstorage.BlobContainer, err error) {

	blobContainerResp, err := blobContainersClient.Get(ctx, resourceGroupName, storageAccountName, containerName, nil)
	if err != nil {
		return
	}
	blobContainer = &blobContainerResp.BlobContainer
	return
}
