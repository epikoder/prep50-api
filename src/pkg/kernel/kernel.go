package kernel

import (
	"fmt"
	"reflect"
)

var typeRegistry = make(map[string]reflect.Type)
var aliasRegistry = make(map[string]string)

func RegisterType(typedNil interface{}, alias string) {
	_, exist := aliasRegistry[alias]
	if exist {
		panic(fmt.Sprintf("could not register type:: alias '%s' already exist", alias))
	}
	t := reflect.TypeOf(typedNil).Elem()
	typeName := t.PkgPath() + "." + t.Name()
	typeRegistry[typeName] = t
	aliasRegistry[alias] = typeName
}

func Make(name string) interface{} {
	return reflect.New(typeRegistry[aliasRegistry[name]]).Interface()
}
