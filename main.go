package main

import (
	"fmt"
	"os"
	"will-translate-cli/config"
	"will-translate-cli/fileops"
	"will-translate-cli/utils"
)

func showHelp() {
    utils.ShowHelp()
}

func main() {
    if len(os.Args) < 2 {
        showHelp()
        return
    }

    command := os.Args[1]

    if command == "--help" {
        showHelp()
        return
    }

	if command == "init" {
		config.CreateConfigFile()
		return
	}

    // Read and parse the configuration file
    configFilePath := "translate.config.json"
    configFileData, err := os.ReadFile(configFilePath)
    if err != nil {
        fmt.Println("Error reading configuration file:", err)
        return
    }

    var cfg config.Configuration
    err = config.UnmarshalConfig(configFileData, &cfg)
    if err != nil {
        fmt.Println("Error parsing configuration file:", err)
        return
    }

    switch command {
    case "create":
        if len(os.Args) < 3 {
            fmt.Println("Usage: create <file1> <file2> ...")
            return
        }
        filenames := os.Args[2:]
        err = fileops.CreateFiles(cfg, filenames...)
        if err != nil {
            fmt.Printf("Error: %v\n", err)
        }

    case "sync":
        fileops.SyncFiles(cfg)

    default:
        fmt.Println("Unknown command. Usage: <create|sync> [<args>]")
    }
}
