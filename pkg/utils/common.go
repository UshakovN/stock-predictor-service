package utils

import (
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
