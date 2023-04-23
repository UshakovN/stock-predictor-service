package utils

import (
  "fmt"
  "reflect"
  "strconv"
)

// CheckRequiredFields check fields with `required: true` tag
//
//  anyStruct pointer to any struct
func CheckRequiredFields(anyStruct any) error {
  if !checkStruct(anyStruct) {
    return fmt.Errorf("specified not a struct pointer")
  }
  const (
    tagKey = "required"
  )
  refVal, refType, err := getStructReflection(anyStruct)
  if err != nil {
    return fmt.Errorf("cannot get struct reflection: %v", err)
  }
  for fieldIdx := 0; fieldIdx < refVal.NumField(); fieldIdx++ {
    refField := refVal.Field(fieldIdx)
    tagVal, tagExist := refType.Field(fieldIdx).Tag.Lookup(tagKey)
    if !tagExist {
      continue
    }
    if trueTagVal, _ := strconv.ParseBool(tagVal); !trueTagVal {
      continue
    }
    fieldName := refType.Field(fieldIdx).Name
    hasValue := refField.IsValid() && !refField.IsZero()
    if !hasValue {
      return fmt.Errorf("field '%s' with tag '%s' is empty", fieldName, tagKey)
    }
    refFieldIface := refField.Interface()

    if checkStruct(refFieldIface) {
      return CheckRequiredFields(refFieldIface)
    }
  }
  return nil
}

// SetDefaultStringValues set default value for string fields in any struct
//
//  anyStruct pointer to any struct
//  defaultValue string value to set
func SetDefaultStringValues(anyStruct any, defaultValue string) error {
  refVal, _, err := getStructReflection(anyStruct)
  if err != nil {
    return fmt.Errorf("cannot get struct reflection: %v", err)
  }
  for fieldIdx := 0; fieldIdx < refVal.NumField(); fieldIdx++ {
    field := refVal.Field(fieldIdx)
    if field.Type().Kind() != reflect.String || !field.CanSet() || field.String() != "" {
      continue
    }
    field.SetString(defaultValue)
  }
  return nil
}

// getStructReflection return reflect value and type for any struct
//
//  anyStruct pointer to any struct
func getStructReflection(anyStruct any) (reflect.Value, reflect.Type, error) {
  refVal := reflect.ValueOf(anyStruct).Elem()
  refType := reflect.TypeOf(anyStruct).Elem()
  if refType.Kind() != reflect.Struct {
    return reflect.Value{}, nil, fmt.Errorf("%v not a struct", anyStruct)
  }
  return refVal, refType, nil
}

func checkStruct(anyStruct any) bool {
  refValue := reflect.ValueOf(anyStruct)
  if refValue.Kind() == reflect.Ptr {
    if refValue.Elem().Kind() == reflect.Struct {
      return true
    }
  }
  return false
}
