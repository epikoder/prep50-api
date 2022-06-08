package list

import (
	"reflect"
)

func Contains(arr interface{}, el interface{}) bool {
	switch reflect.TypeOf(arr).Kind() {
	case reflect.Slice:
		{
			v := reflect.ValueOf(arr)
			for i := 0; i < v.Len(); i++ {
				if v.Index(i).Interface() == el {
					return true
				}
			}
			return false
		}
	default:
		return false
	}
}
