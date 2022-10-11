package list

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"

	"github.com/Prep50mobileApp/prep50-api/src/pkg/dbmodel"
	"github.com/google/uuid"
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

func DeleteUint(s []uint, index int) []uint {
	ret := make([]uint, 0)
	ret = append(ret, s[:index]...)
	return append(ret, s[index+1:]...)
}

func Slug(s string) string {
	_s := ""
	for _, r := range strings.ReplaceAll(strings.TrimSpace(s), "	", " ") {
		if unicode.IsLetter(r) {
			_s = _s + string(r)
			continue
		}
		if unicode.IsDigit(r) {
			_s = _s + string(r)
			continue
		}
		_s = strings.TrimSuffix(_s, "-")
		_s = _s + "-"
	}
	return strings.ToLower(strings.TrimSuffix(_s, "-"))
}

func UniqueSlug(model dbmodel.DBModel, slug string) (s string, err error) {
	var i = 1
	for {
		s = fmt.Sprintf("%s%d", slug, i)
		m := reflect.New(reflect.TypeOf(model).Elem()).Interface().(dbmodel.DBModel)
		if err = model.Database().Find(m, "slug = ?", s).Error; err != nil {
			return
		}
		uid, ok := m.ID().(uuid.UUID)
		if !ok {
			if id := m.ID().(int); id == 0 {
				return
			}
		} else if uid == uuid.Nil {
			return
		}
		i++
	}
}

func UniqueByField(model dbmodel.DBModel, value, field string) (s string, err error) {
	var i = 1
	for {
		s = fmt.Sprintf("%s%d", value, i)
		m := reflect.New(reflect.TypeOf(model).Elem()).Interface().(dbmodel.DBModel)
		if err = model.Database().Find(m, "? = ?", field, s).Error; err != nil {
			return
		}
		uid, ok := m.ID().(uuid.UUID)
		if !ok {
			if id := m.ID().(int); id == 0 {
				return
			}
		} else if uid == uuid.Nil {
			return
		}
		i++
	}
}
