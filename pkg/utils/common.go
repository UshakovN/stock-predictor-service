package utils

import (
  "encoding/json"
  "fmt"
  "strings"
)

func ExtractOptional[T any](optional ...T) T {
  var value T

  for _, val := range optional {
    value = val
  }
  return value
}

func TitleString(s string) string {
  return strings.Title(strings.ToLower(s))
}

func StripString(s string) string {
  return strings.TrimSpace(s)
}

func ExtractFileExtension(filePath string) (string, error) {
  const (
    minPartsCount   = 2
    maxExtensionLen = 10
  )
  partsURL := strings.Split(filePath, ".")
  partsCount := len(partsURL)

  if partsCount < minPartsCount {
    return "", fmt.Errorf("image extension not found")
  }
  imageExtension := partsURL[partsCount-1]

  if len([]rune(imageExtension)) > maxExtensionLen {
    return "", fmt.Errorf("malformed image extension '%s'", imageExtension)
  }
  return imageExtension, nil
}

func ToMap[T any, K comparable](items []T, extractor func(T) K) map[K]T {
  itemsMap := map[K]T{}

  for _, item := range items {
    itemsMap[extractor(item)] = item
  }
  return itemsMap
}

func FillFrom[T, S any](source T, dest S) error {
  bytes, err := json.Marshal(source)
  if err != nil {
    return fmt.Errorf("marshal failed: %v", err)
  }
  if err = json.Unmarshal(bytes, dest); err != nil {
    return fmt.Errorf("unmarshal failed: %v", err)
  }
  return nil
}
