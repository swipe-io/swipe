package option

import (
	"fmt"
	"reflect"
)

func Encode(i interface{}) error {

	return encode(i)
}

func encode(v interface{}) error {

	t := reflect.TypeOf(v)
	o := reflect.ValueOf(v)

	if t.Kind() == reflect.Ptr {
		t = t.Elem() //Gets the type in the type pointer
	}
	if o.Kind() == reflect.Ptr {
		o = o.Elem() //Get the value in the value address
	}

	for i := 0; i < t.NumField(); i++ {
		f := o.Field(i)

		fmt.Println(f.Kind(), reflect.TypeOf(f.Interface()).Elem().Kind())
	}

	//valueOf := reflect.Indirect(reflect.ValueOf(i))

	//for i := 0; i < v.NumField(); i++ {
	//	v := v.Field(i)
	//	t := v.Type().Field(i).Type
	//
	//	if t.Kind() == reflect.Ptr {
	//		t = t.Elem() //Gets the type in the type pointer
	//	}
	//	if v.Kind() == reflect.Ptr {
	//		v = v.Elem() //Get the value in the value address
	//	}
	//
	//	fmt.Println(v, t)
	//	//varName := v.Type().Field(i).Name
	//	//varType := v.Type().Field(i).Type
	//	//varValue := v.Field(i).Interface()
	//	//varTag := v.Type().Field(i).Tag
	//
	//	//if name := varTag.Get("mapstructure"); name != "" {
	//	//	fmt.Println(name)
	//	//} else {
	//	//	fmt.Println(varName)
	//	//}
	//	//
	//	//if v.Field(i).Kind() == reflect.Struct {
	//	//	if err := encode(v.Field(i)); err != nil {
	//	//		return err
	//	//	}
	//	//	continue
	//	//}
	//
	//}
	return nil
}
