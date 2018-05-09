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

func getTagValues(check string, tagString string) []string {
	stringSplit := strings.Split(tagString, ";")
	for _, tag := range stringSplit {
		tagSplit := strings.Split(tag, ":")
		if len(tagSplit) != 2 {
			return []string{}
		}

		if tagSplit[0] == check {
			return strings.Split(tagSplit[1], ",")
		}
	}

	return []string{}
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

func UpperCamelCase(s string) string {
	return camelCase(s, true)
}

func LowerCamelCase(s string) string {
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
