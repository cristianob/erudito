package erudito

import (
	"reflect"
	"strings"
)

func checkIfTagExists(check string, tags string) bool {
	stringSplit := strings.Split(tags, ";")
	for _, tag := range stringSplit {
		if tag == check {
			return true
		}
	}

	return false
}

func deepCopy(model reflect.Type, source, destination reflect.Value, excludeTag string) {
	for i := 0; i < model.NumField(); i++ {

		switch model.Field(i).Type.Kind() {
		case reflect.Struct:
			deepCopy(source.Field(i).Type(), source.Field(i), destination.Field(i), excludeTag)

		default:
			if !checkIfTagExists(excludeTag, model.Field(i).Tag.Get("erudito")) {
				if len(model.Field(i).PkgPath) != 0 {
					continue
				}

				destination.Field(i).Set(source.Field(i))
			}
		}
	}
}
