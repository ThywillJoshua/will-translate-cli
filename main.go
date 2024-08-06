package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

// Configuration holds the structure of the JSON configuration
type Configuration struct {
	ProjectName string `json:"project-name"`
	Config      Config `json:"config"`
}

// Config holds the nested configuration details
type Config struct {
	FilesPath   string `json:"files-path"`
	DefaultFile string `json:"default-file"`
}

// SortedMap holds the sorted key-value pairs
type SortedMap struct {
	Keys   []string
	Values map[string]interface{}
}

// UnmarshalJSON custom unmarshaller for SortedMap
func (sm *SortedMap) UnmarshalJSON(data []byte) error {
	var temp map[string]interface{}
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	sm.Values = make(map[string]interface{})
	sm.Keys = make([]string, 0, len(temp))

	for key := range temp {
		sm.Keys = append(sm.Keys, key)
	}
	sort.Strings(sm.Keys)

	for _, key := range sm.Keys {
		sm.Values[key] = temp[key]
	}
	return nil
}

// MarshalJSON custom marshaller for SortedMap
func (sm SortedMap) MarshalJSON() ([]byte, error) {
	m := make(map[string]interface{})
	for _, key := range sm.Keys {
		m[key] = sm.Values[key]
	}
	return json.Marshal(m)
}

// sortSortedMap returns a SortedMap with keys sorted alphabetically
func sortSortedMap(sm SortedMap) SortedMap {
	sortedMap := SortedMap{
		Keys:   make([]string, len(sm.Keys)),
		Values: make(map[string]interface{}),
	}
	copy(sortedMap.Keys, sm.Keys)
	sort.Strings(sortedMap.Keys)

	for _, key := range sortedMap.Keys {
		sortedMap.Values[key] = sm.Values[key]
	}
	return sortedMap
}

// updateSortedMap updates the content of the target map with the default map
func updateSortedMap(target *SortedMap, defaultContent SortedMap, pathPrefix string) (int, []string, int, []string) {
	defaultMap := make(map[string]interface{})
	for _, key := range defaultContent.Keys {
		defaultMap[key] = defaultContent.Values[key]
	}

	var removedKeys []string
	var removalCount int
	var addedKeys []string
	var additionCount int

	// Iterate over the default keys to add or update target keys
	for _, key := range defaultContent.Keys {
		fullPath := pathPrefix + key
		if targetValue, exists := target.Values[key]; !exists {
			// Key exists in the default but not in the target file
			addedKeys = append(addedKeys, fullPath)
			target.Values[key] = defaultMap[key]
			target.Keys = append(target.Keys, key)
			additionCount++
		} else {
			// Recursively update nested objects
			if targetMap, ok := targetValue.(map[string]interface{}); ok {
				if defaultMapValue, ok := defaultMap[key].(map[string]interface{}); ok {
					// Convert nested maps to SortedMap
					defaultSortedMap := mapToSortedMap(defaultMapValue)
					targetSortedMap := mapToSortedMap(targetMap)

					// Recursive update
					nestedRemovalCount, nestedRemovedKeys, nestedAdditionCount, nestedAddedKeys := updateSortedMap(&targetSortedMap, defaultSortedMap, fullPath+".")
					if nestedRemovalCount > 0 || nestedAdditionCount > 0 {
						target.Values[key] = targetSortedMap.Values
						removalCount += nestedRemovalCount
						removedKeys = append(removedKeys, nestedRemovedKeys...)
						additionCount += nestedAdditionCount
						addedKeys = append(addedKeys, nestedAddedKeys...)
					}
				}
			}
		}
	}

	// Remove keys that are in target but not in default
	for _, key := range target.Keys {
		if _, exists := defaultMap[key]; !exists {
			fullPath := pathPrefix + key
			removedKeys = append(removedKeys, fullPath)
			delete(target.Values, key)
			target.Keys = removeElement(target.Keys, key)
			removalCount++
		}
	}

	// Sort the keys again after updating
	sort.Strings(target.Keys)
	return removalCount, removedKeys, additionCount, addedKeys
}

// mapToSortedMap converts a map to a SortedMap
func mapToSortedMap(m map[string]interface{}) SortedMap {
	sm := SortedMap{
		Keys:   make([]string, 0, len(m)),
		Values: m,
	}
	for k := range m {
		sm.Keys = append(sm.Keys, k)
	}
	sort.Strings(sm.Keys)
	return sm
}

// removeElement removes an element from a slice of strings
func removeElement(slice []string, element string) []string {
	for i, v := range slice {
		if v == element {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

// createFiles creates files with empty JSON objects in the specified directory
func createFiles(config Configuration, filenames ...string) error {
	for _, filename := range filenames {
		filePath := filepath.Join(config.Config.FilesPath, filename)

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

func syncFiles(config Configuration) {
	// Construct the path to the default file
	defaultFilePath := filepath.Join(config.Config.FilesPath, config.Config.DefaultFile)

	// Read the JSON default file
	defaultFileData, err := os.ReadFile(defaultFilePath)
	if err != nil {
		fmt.Println("Error reading default file:", err)
		return
	}

	// Parse the JSON default file content into SortedMap and sort it
	var defaultFileContent SortedMap
	err = json.Unmarshal(defaultFileData, &defaultFileContent)
	if err != nil {
		fmt.Println("Error parsing default file:", err)
		return
	}
	defaultFileContent = sortSortedMap(defaultFileContent)

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
	err = filepath.WalkDir(config.Config.FilesPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Println("Error walking through directory:", err)
			return err
		}

		// Process only files that end with .json and are not the default file
		if filepath.Ext(d.Name()) == ".json" && d.Name() != config.Config.DefaultFile {
			// Read the target file
			targetFileData, err := os.ReadFile(path)
			if err != nil {
				fmt.Printf("Error reading file %s: %v\n", path, err)
				return nil
			}

			// Parse the target file content into SortedMap
			var targetFileContent SortedMap
			err = json.Unmarshal(targetFileData, &targetFileContent)
			if err != nil {
				fmt.Printf("Error parsing file %s: %v\n", path, err)
				return nil
			}

			// Update the target file content with the default content
			removalCount, removedKeys, additionCount, addedKeys := updateSortedMap(&targetFileContent, defaultFileContent, "")
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

func createConfigFile() {
	// Define the filename and content
	filename := "translate.config.json"
	content := map[string]interface{}{
		"project-name": "<name of project>",
		"config": map[string]string{
			"files-path":   "<path of translation files>",
			"default-file": "<path of default file>",
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

// showHelp displays usage information for the application
func showHelp() {
	fmt.Println("Usage: translate <command> [<args>]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  create <file1> <file2> ...      Create new JSON files in the directory specified in the configuration.")
	fmt.Println("                                 Example: translate create en.json fr.json")
	fmt.Println()
	fmt.Println("  sync                           Sync all JSON files in the directory with the default file specified in the configuration.")
	fmt.Println("                                 This will add missing keys and remove extra keys to match the default file.")
	fmt.Println("                                 Example: translate sync")
	fmt.Println()
	fmt.Println("  init                           Create the initial configuration file 'translate.config.json'.")
	fmt.Println("                                 This command should be run before any other commands if the config file does not exist.")
	fmt.Println("                                 Example: translate init")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --help                         Show this help message.")
	fmt.Println()
	fmt.Println("Description:")
	fmt.Println("  The 'translate' tool helps in managing and synchronizing translation JSON files.")
	fmt.Println("  The path for the 'create' and 'sync' commands is obtained from the configuration file 'translate.config.json'.")
	fmt.Println("  The 'create' command generates new JSON files with empty objects.")
	fmt.Println("  The 'sync' command updates existing JSON files with the content of the default JSON file,")
	fmt.Println("  ensuring that all files have consistent keys.")
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

	// Read the configuration file
	configFilePath := "translate.config.json"
	configFileData, err := os.ReadFile(configFilePath)
	if err != nil {
		fmt.Println("Error reading configuration file:", err)
		return
	}

	// Parse the configuration file
	var config Configuration
	err = json.Unmarshal(configFileData, &config)
	if err != nil {
		fmt.Println("Error parsing configuration file:", err)
		return
	}

	switch command {
	case "init":
		createConfigFile()

	case "create":
		if len(os.Args) < 3 {
			fmt.Println("Usage: create <file1> <file2> ...")
			return
		}
        // Parse the filenames
		args := os.Args[2:]

		// Call the createFiles function
		err = createFiles(config, args...)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}

	case "sync":
		// Sync the files with the default content
		syncFiles(config)

	default:
		fmt.Println("Unknown command. Usage: <create|sync> [<args>]")
	}
}
