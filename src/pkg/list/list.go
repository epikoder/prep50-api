package list

import (
	"reflect"
	"strings"
)

func Contains(arr interface{}, el interface{}) bool {
	switch reflect.TypeOf(arr).Kind() {
	case reflect.Slice:
		{
			v := reflect.ValueOf(arr)
			switch el.(type) {
			case string:
				{
					for i := 0; i < v.Len(); i++ {
						val, ok := v.Index(i).Interface().(string)
						if !ok {
							return false
						}
						val2, ok := el.(string)
						if !ok {
							return false
						}

						if strings.EqualFold(val, val2) {
							return true
						}
					}
					return false
				}
			default:
				for i := 0; i < v.Len(); i++ {
					if v.Index(i).Interface() == el {
						return true
					}
				}
				return false
			}
		}
	default:
		return false
	}
}

func Unique(arr interface{}) interface{} {
	switch reflect.TypeOf(arr).Kind() {
	case reflect.Slice | reflect.Array:
		{
			v := reflect.ValueOf(arr)
			if v.Len() == 0 {
				return nil
			}
			keys := make(map[interface{}]bool)
			switch v.Index(0).Interface().(type) {
			case int, int16, int32, int64:
				{
					list := []int{}
					for i := 0; i < v.Len(); i++ {
						val := v.Index(i).Interface()
						if _, ok := keys[val]; !ok {
							keys[val] = true
							list = append(list, val.(int))
						}
					}
					return list
				}
			case uint, uint16, uint32, uint64:
				{
					list := []uint{}
					for i := 0; i < v.Len(); i++ {
						val := v.Index(i).Interface()
						if _, ok := keys[val]; !ok {
							keys[val] = true
							list = append(list, val.(uint))
						}
					}
					return list
				}
			default:
				list := []interface{}{}
				for i := 0; i < v.Len(); i++ {
					val := v.Index(i).Interface()
					if _, ok := keys[val]; !ok {
						keys[val] = true
						list = append(list, val)
					}
				}
				return list
			}
		}
	default:
		return nil
	}
}

func DeleteInt(s []int, index int) []int {
	ret := make([]int, 0)
	ret = append(ret, s[:index]...)
	return append(ret, s[index+1:]...)
}
