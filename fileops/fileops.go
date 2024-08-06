package fileops

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"will-translate-cli/config"
	"will-translate-cli/sortedmap"
)

func CreateFiles(cfg config.Configuration, filenames ...string) error {
    for _, filename := range filenames {
        filePath := filepath.Join(cfg.Config.FilesPath, filename)

        // Create the file and write an empty JSON object
        file, err := os.Create(filePath)
        if err != nil {
            return fmt.Errorf("error creating file %s: %v", filePath, err)
        }
        defer file.Close()

        _, err = file.WriteString("{}")
        if err != nil {
            return fmt.Errorf("error writing to file %s: %v", filePath, err)
        }

        fmt.Printf("Created file: %s\n", filePath)
    }

    return nil
}

func SyncFiles(cfg config.Configuration) {
    // Construct the path to the default file
    defaultFilePath := filepath.Join(cfg.Config.FilesPath, cfg.Config.DefaultFile)

    // Read the JSON default file
    defaultFileData, err := os.ReadFile(defaultFilePath)
    if err != nil {
        fmt.Println("Error reading default file:", err)
        return
    }

    // Parse the JSON default file content into SortedMap and sort it
    var defaultFileContent sortedmap.SortedMap
    err = json.Unmarshal(defaultFileData, &defaultFileContent)
    if err != nil {
        fmt.Println("Error parsing default file:", err)
        return
    }
    defaultFileContent = sortedmap.SortSortedMap(defaultFileContent)

    // Convert the sorted default file content back to JSON
    defaultFileContentIndented, err := json.MarshalIndent(defaultFileContent, "", "  ")
    if err != nil {
        fmt.Println("Error marshalling sorted default file content:", err)
        return
    }

    // Write the sorted default file content back to the default file
    err = os.WriteFile(defaultFilePath, defaultFileContentIndented, 0644)
    if err != nil {
        fmt.Println("Error writing sorted default file:", err)
        return
    }

    // Get all JSON files in the specified directory
    err = filepath.WalkDir(cfg.Config.FilesPath, func(path string, d fs.DirEntry, err error) error {
        if err != nil {
            fmt.Println("Error walking through directory:", err)
            return err
        }

        // Process only files that end with .json and are not the default file
        if filepath.Ext(d.Name()) == ".json" && d.Name() != cfg.Config.DefaultFile {
            // Read the target file
            targetFileData, err := os.ReadFile(path)
            if err != nil {
                fmt.Printf("Error reading file %s: %v\n", path, err)
                return nil
            }

            // Parse the target file content into SortedMap
            var targetFileContent sortedmap.SortedMap
            err = json.Unmarshal(targetFileData, &targetFileContent)
            if err != nil {
                fmt.Printf("Error parsing file %s: %v\n", path, err)
                return nil
            }

            // Update the target file content with the default content
            removalCount, removedKeys, additionCount, addedKeys := sortedmap.UpdateSortedMap(&targetFileContent, defaultFileContent, "")
            if removalCount > 0 {
                fmt.Printf("Removed %d keys from file: %s\n", removalCount, path)
                for _, key := range removedKeys {
                    fmt.Printf("Removed key: %s\n", key)
                }
            }
            if additionCount > 0 {
                fmt.Printf("Added %d keys to file: %s\n", additionCount, path)
                for _, key := range addedKeys {
                    fmt.Printf("Added key: %s\n", key)
                }
            }
            if removalCount == 0 && additionCount == 0 {
                fmt.Printf("No changes made to file: %s\n", path)
            }

            // Convert the updated target file content back to JSON
            targetFileContentIndented, err := json.MarshalIndent(targetFileContent, "", "  ")
            if err != nil {
                fmt.Printf("Error marshalling updated file content for %s: %v\n", path, err)
                return nil
            }

            // Write the updated target file content back to the file
            err = os.WriteFile(path, targetFileContentIndented, 0644)
            if err != nil {
                fmt.Printf("Error writing updated file content for %s: %v\n", path, err)
                return nil
            }

            fmt.Printf("Updated file: %s\n", path)
        }
        return nil
    })

    if err != nil {
        fmt.Println("Error walking through directory:", err)
    }
}
