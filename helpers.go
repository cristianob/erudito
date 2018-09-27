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

func getJSONFieldNameByTag(tag string) string {
	stringSplit := strings.Split(tag, ",")
	return stringSplit[0]
}

func fieldIsAEruditoModel(field reflect.StructField) bool {
	switch field.Type.Kind() {
	case reflect.Slice:
		return field.Type.Elem().Implements(reflect.TypeOf((*Model)(nil)).Elem()) ||
			field.Type.Elem() == reflect.TypeOf(FullModel{}) ||
			field.Type.Elem() == reflect.TypeOf(HardDeleteModel{}) ||
			field.Type.Elem() == reflect.TypeOf(SimpleModel{})
	case reflect.Struct:
		return field.Type.Implements(reflect.TypeOf((*Model)(nil)).Elem()) ||
			field.Type == reflect.TypeOf(FullModel{}) ||
			field.Type == reflect.TypeOf(HardDeleteModel{}) ||
			field.Type == reflect.TypeOf(SimpleModel{})
	default:
		return false
	}
}

func upperCamelCase(s string) string {
	return camelCase(s, true)
}

func lowerCamelCase(s string) string {
	return camelCase(s, false)
}

func camelCase(s string, upper bool) string {
	s = strings.TrimSpace(s)
	buffer := make([]rune, 0, len(s))

	var prev rune
	for _, curr := range s {
		if curr != '_' {
			if prev == '_' || (upper && prev == 0) || prev == '.' {
				buffer = append(buffer, toUpper(curr))
			} else {
				buffer = append(buffer, toLower(curr))
			}
		}
		prev = curr
	}

	return string(buffer)
}

func toUpper(ch rune) rune {
	if ch >= 'a' && ch <= 'z' {
		return ch - 32
	}
	return ch
}

func toLower(ch rune) rune {
	if ch >= 'A' && ch <= 'Z' {
		return ch + 32
	}
	return ch
}
