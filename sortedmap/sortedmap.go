package sortedmap

import (
	"encoding/json"
	"sort"
)

type SortedMap struct {
    Keys   []string
    Values map[string]interface{}
}

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

func (sm SortedMap) MarshalJSON() ([]byte, error) {
    m := make(map[string]interface{})
    for _, key := range sm.Keys {
        m[key] = sm.Values[key]
    }
    return json.Marshal(m)
}

func SortSortedMap(sm SortedMap) SortedMap {
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

func UpdateSortedMap(target *SortedMap, defaultContent SortedMap, pathPrefix string) (int, []string, int, []string) {
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
                    defaultSortedMap := MapToSortedMap(defaultMapValue)
                    targetSortedMap := MapToSortedMap(targetMap)

                    // Recursive update
                    nestedRemovalCount, nestedRemovedKeys, nestedAdditionCount, nestedAddedKeys := UpdateSortedMap(&targetSortedMap, defaultSortedMap, fullPath+".")
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
            target.Keys = RemoveElement(target.Keys, key)
            removalCount++
        }
    }

    // Sort the keys again after updating
    sort.Strings(target.Keys)
    return removalCount, removedKeys, additionCount, addedKeys
}

func MapToSortedMap(m map[string]interface{}) SortedMap {
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

func RemoveElement(slice []string, element string) []string {
    for i, v := range slice {
        if v == element {
            return append(slice[:i], slice[i+1:]...)
        }
    }
    return slice
}
