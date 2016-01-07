package main

import (
	"fmt"
	"github.com/codegangsta/cli"
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
		storageAccount := os.Args[1]
		container := os.Args[2]
		source := os.Args[3]
		fmt.Printf("Uploading %s to %s/%s\n", source, storageAccount, container)
	}
	app.Run(os.Args)
}
