package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/sas"
)

var (
	subscriptionID     string
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

	accountKey := os.Getenv("AZURE_ACCOUNT_KEY")
	if len(accountKey) == 0 {
		log.Fatal("AZURE_ACCOUNT_KEY is not set.")
	}
	scred, err := azblob.NewSharedKeyCredential(storageAccountName, accountKey)
	blob_client, err := azblob.NewClientWithSharedKeyCredential(fmt.Sprintf("https://%s.blob.core.windows.net/", storageAccountName), scred, nil)
	sas_url, err := blob_client.ServiceClient().GetSASURL(
		sas.AccountResourceTypes{ Container: true },
		sas.AccountPermissions{
			Create: true, Delete: true, List: true, Add: true,
		},
		time.Now().Add(24 * time.Hour),
		nil,	
	)
	if err != nil {
		log.Fatal(err)
		return
	}
	fmt.Printf("SAS URL %v\n", sas_url)

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

			currentTime := time.Now()
			tomorrow := currentTime.Add(24 * time.Hour)

			client := storageClientFactory.NewAccountsClient()
			sasToken, err := client.ListAccountSAS(ctx, resourceGroupName, storageAccountName, armstorage.AccountSasParameters{
				KeyToSign:              to.Ptr("key1"),
				SharedAccessExpiryTime: to.Ptr(tomorrow.Round(time.Microsecond).UTC()),
				Permissions:            to.Ptr(armstorage.PermissionsR),
				Protocols:              to.Ptr(armstorage.HTTPProtocolHTTPSHTTP),
				ResourceTypes:          to.Ptr(armstorage.SignedResourceTypesS),
				Services:               to.Ptr(armstorage.ServicesB),
				SharedAccessStartTime:  to.Ptr(currentTime.Round(time.Microsecond).UTC()),
			}, nil)
			if err != nil {
				log.Fatal(err)
				return
			}
			token := *sasToken.AccountSasToken
			fmt.Printf("SAS Token %v\n", token)
			return
		} else {
			fmt.Printf("Container Get status code %d error code: %v", respErr.StatusCode, respErr.ErrorCode)
			log.Fatal(respErr)
		}
	} else {
		id := *blobContainer.ID
		fmt.Printf("Blob container already exists, id: %v", id)
	}
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
