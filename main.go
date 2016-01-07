package main

import (
	"encoding/base64"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/storage"
	"github.com/codegangsta/cli"
	"io"
	"os"
	"path/filepath"
)

func walkFiles(source string, ignore string) ([]string, error) {
	var items []string
	err := filepath.Walk(source, func(p string, fi os.FileInfo, err error) error {
		if err != nil || fi.IsDir() {
			return err
		}
		rel, err := filepath.Rel(source, p)
		if err != nil {
			return err
		}

		var ignoreFile bool
		if ignore != "" {
			ignoreFile, err = filepath.Match(ignore, rel)
		}
		if err != nil || ignoreFile {
			return err
		}
		items = append(items, p)
		return nil
	})
	return items, err
}

func main() {
	app := cli.NewApp()
	app.Name = "blobxfer"
	app.Usage = "Upload files to Azure Blob Storage"
	app.Author = "Matthew Brown <me@matthewbrown.io>"
	app.Version = "0.1.0"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "storageaccountkey",
			Value:  "",
			Usage:  "The storage account key to use",
			EnvVar: "STORAGE_ACCOUNT_KEY",
		},
	}
	app.Action = func(ctx *cli.Context) {

		if len(os.Args) != 4 {
			fmt.Errorf("blobxfer requires 3 command line arguments. Got %v", len(os.Args))
			os.Exit(1)
		}

		storageAccount := os.Args[1]
		container := os.Args[2]
		source := os.Args[3]

		fmt.Printf("Uploading %s to %s/%s\n", source, storageAccount, container)
		filelist, err := walkFiles(source, "")
		if err != nil {
			fmt.Errorf("Failed to open source: %s", err)
		}

		storageClient, err := storage.NewBasicClient(storageAccount, "")
		if err != nil {
			fmt.Errorf("Failed to create client: %s", err)
		}
		blobClient := storageClient.GetBlobService()

		for _, f := range filelist {
			blobURI := blobClient.GetBlobURL(container, f)
			fmt.Printf("Uploading: %s\n", blobURI)
			fd, err := os.Open(f)
			if err != nil {
				fmt.Errorf("Error reading file %s: %s", f, err)
				os.Exit(1)
			}
			defer fd.Close()

			err = blobClient.CreateBlockBlob(container, f)
			if err != nil {
				fmt.Errorf("Failed to create blobblock: %s", err)
				os.Exit(1)
			}

			err = putBlockBlob(blobClient, container, f, fd, storage.MaxBlobBlockSize)
			if err != nil {
				fmt.Errorf("Failed to put blob: %v", err)
				os.Exit(1)
			}
		}

	}
	app.Run(os.Args)
}

func putBlockBlob(b storage.BlobStorageClient, container, name string, blob io.Reader, chunkSize int) error {
	if chunkSize <= 0 || chunkSize > storage.MaxBlobBlockSize {
		chunkSize = storage.MaxBlobBlockSize
	}

	chunk := make([]byte, chunkSize)
	n, err := blob.Read(chunk)
	if err != nil && err != io.EOF {
		return err
	}

	blockList := []storage.Block{}

	for blockNum := 0; ; blockNum++ {
		id := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%011d", blockNum)))
		data := chunk[:n]
		err = b.PutBlock(container, name, id, data)
		if err != nil {
			return err
		}

		blockList = append(blockList, storage.Block{ID: id, Status: storage.BlockStatusLatest})

		// Read next block
		n, err = blob.Read(chunk)
		if err != nil && err != io.EOF {
			return err
		}
		if err == io.EOF {
			break
		}
	}

	return b.PutBlockList(container, name, blockList)
}
