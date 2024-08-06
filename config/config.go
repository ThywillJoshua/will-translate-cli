package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Configuration struct {
    ProjectName string `json:"project-name"`
    Config      Config `json:"config"`
}

type Config struct {
    FilesPath   string `json:"files-path"`
    DefaultFile string `json:"default-file"`
}

func UnmarshalConfig(data []byte, cfg *Configuration) error {
    return json.Unmarshal(data, cfg)
}

func CreateConfigFile() {
    // Define the filename and content
    filename := "translate.config.json"
    content := map[string]interface{}{
        "project-name": "my-project",
        "config": map[string]string{
            "files-path":   "public/i18n",
            "default-file": "gb-en.json",
        },
    }

    // Check if the file already exists
    if _, err := os.Stat(filename); err == nil {
        fmt.Printf("The file %s already exists. Please delete it first and run the command again.\n", filename)
        return
    } else if !os.IsNotExist(err) {
        fmt.Println("Error checking file:", err)
        return
    }

    // Convert content to JSON
    fileContent, err := json.MarshalIndent(content, "", "  ")
    if err != nil {
        fmt.Println("Error marshalling JSON:", err)
        return
    }

    // Write the content to the config file
    err = os.WriteFile(filename, fileContent, 0644)
    if err != nil {
        fmt.Println("Error writing file:", err)
        return
    }

    fmt.Printf("%s has been created with the initial configuration.\n", filename)
}
