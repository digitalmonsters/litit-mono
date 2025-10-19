package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/sas"
	"github.com/digitalmonsters/go-common/azure_blob"
	"github.com/digitalmonsters/go-common/boilerplate"
)

func init() {
	initConfigs()
}

var configuration *boilerplate.AzureBlobConfig

func main() {
	PutRequest()
	GetRequest()
	GetBlobMetadataRequest()
	UploadRequest()
}

func initConfigs() {

	configuration = &boilerplate.AzureBlobConfig{
		UseCliAuth:         false,
		StorageAccountName: "testaccountname",
		StorageAccountKey:  "testkey",
		Container:          "testcontainer",
	}

	raw, _ := json.Marshal(configuration)
	fmt.Println("configurations:", string(raw))
}

func PutRequest() {
	client := azure_blob.NewAzureBlobObject(configuration)
	fmt.Println("azure client successfull")

	url, err := client.PutObjectSignedUrl("test/internal/config_1.json", 15*time.Hour, "")
	if err != nil {
		log.Println("failed to PutObjectSignedUrl blob - " + err.Error())
		return
	}

	fmt.Println("url", url)

	parts, err := sas.ParseURL(url)
	if err != nil {
		log.Println("failed to parse url - " + err.Error())
		return
	}

	fmt.Println("parsed url: ", parts.ContainerName, parts.BlobName, parts.Host, parts.SAS.Permissions())

	req, err := http.NewRequest(http.MethodPut, url, strings.NewReader("Hello World!!"))
	if err != nil {
		log.Println("http req failed - " + err.Error())
		return
	}
	req.Header.Add("x-ms-blob-type", "BlockBlob")

	fmt.Println("req header: ", req.Header)
	fmt.Println("req query: ", req.URL.Query())
	fmt.Println("req url: ", req.URL.String())

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("http call failed - " + err.Error())
		return
	}

	fmt.Printf("client: got response!\n")
	fmt.Printf("client: status code: %d\n", res.StatusCode)

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Printf("client: could not read response body: %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("client: response body: \n%s\n", resBody)
}

func GetRequest() {
	client := azure_blob.NewAzureBlobObject(configuration)
	fmt.Println("azure client successfull")

	url, err := client.GetObjectSignedUrl("test/internal/config_1.json", 15*time.Hour)
	if err != nil {
		log.Println("failed to GetObjectSignedUrl blob - " + err.Error())
		return
	}

	fmt.Println("url", url)

	parts, err := sas.ParseURL(url)
	if err != nil {
		log.Println("failed to parse url - " + err.Error())
		return
	}

	fmt.Println("parsed url: ", parts.ContainerName, parts.BlobName, parts.Host, parts.SAS.Permissions())

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Println("http req failed - " + err.Error())
		return
	}
	req.Header.Add("x-ms-blob-type", "BlockBlob")

	fmt.Println("req header: ", req.Header)
	fmt.Println("req query: ", req.URL.Query())
	fmt.Println("req url: ", req.URL.String())

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("http call failed - " + err.Error())
		return
	}

	fmt.Printf("client: got response!\n")
	fmt.Printf("client: status code: %d\n", res.StatusCode)

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Printf("client: could not read response body: %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("client: response body: \n%s\n\n", resBody)
}

func GetBlobMetadataRequest() {
	client := azure_blob.NewAzureBlobObject(configuration)
	fmt.Println("azure client successfull")
	size, err := client.GetObjectSize("test/internal/config_1.json")
	if err != nil {
		log.Println("failed to GetObjectSize blob - " + err.Error())
	}

	fmt.Printf("client: response size in Bytes: %v\n", size)
}

func UploadRequest() {
	client := azure_blob.NewAzureBlobObject(configuration)
	fmt.Println("azure client successfull")
	err := client.UploadObject("test/internal/config_upload.json", []byte("Lots of data!!"), "")
	if err != nil {
		log.Println("failed to GetObjectSize blob - " + err.Error())
	}

	fmt.Printf("client: upload successful")
}
