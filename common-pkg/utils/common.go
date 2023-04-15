package utils

import (
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
