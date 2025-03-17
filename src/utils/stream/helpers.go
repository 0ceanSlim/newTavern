package stream

import (
	"fmt"
	"math/rand"
	"reflect"
)

func generateDtag() string {
	return fmt.Sprintf("%d", rand.Intn(900000)+100000)
}

// Helper function to convert MetadataConfig to map[string]interface{}
func structToMap(obj interface{}) map[string]interface{} {
	objT := reflect.TypeOf(obj)
	objV := reflect.ValueOf(obj)

	if objT.Kind() == reflect.Ptr {
		objT = objT.Elem()
		objV = objV.Elem()
	}

	result := make(map[string]interface{})
	for i := 0; i < objT.NumField(); i++ {
		fieldT := objT.Field(i)
		fieldV := objV.Field(i)
		tag := fieldT.Tag.Get("yaml") // Use yaml tag
		if tag == "" {
			tag = fieldT.Name
		}
		result[tag] = fieldV.Interface()
	}
	return result
}
