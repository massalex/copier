package copier

import (
	"database/sql"
	"errors"
	"reflect"
)

type instance struct {
	isSlice   bool
	amount    int
	from      reflect.Value
	to        reflect.Value
	mapSuffix string
}

func New(from interface{}, to interface{}, mapSuffix string) *instance {
	return &instance{
		amount: 1,
		from: indirect(reflect.ValueOf(from)),
		to: indirect(reflect.ValueOf(to)),
		mapSuffix: mapSuffix,
	}
}

// Copy copy things
func (copier *instance) Copy() (err error) {
	if !copier.to.CanAddr() {
		return errors.New("copy to value is unaddressable")
	}

	// Return is from value is invalid
	if !copier.from.IsValid() {
		return errors.New("from value is invalid")
	}

	// Just set it if possible to assign
	if copier.mapSuffix == "" && copier.from.Type().AssignableTo(copier.to.Type()) {
		copier.to.Set(copier.from)
		return
	}

	fromType := indirectType(copier.from.Type())
	toType := indirectType(copier.to.Type())

	if fromType.Kind() != reflect.Struct || toType.Kind() != reflect.Struct {
		return
	}

	if copier.to.Kind() == reflect.Slice {
		copier.isSlice = true
		if copier.from.Kind() == reflect.Slice {
			copier.amount = copier.from.Len()
		}
	}

	for i := 0; i < copier.amount; i++ {
		var dest, source reflect.Value

		if copier.isSlice {
			// source
			if copier.from.Kind() == reflect.Slice {
				source = indirect(copier.from.Index(i))
			} else {
				source = indirect(copier.from)
			}

			// dest
			dest = indirect(reflect.New(toType).Elem())
		} else {
			source = indirect(copier.from)
			dest = indirect(copier.to)
		}

		// Copy from field to field or map method or method
		for _, field := range deepFields(fromType) {
			name := field.Name

			if fromField := source.FieldByName(name); fromField.IsValid() {
				canAddr := dest.CanAddr()

				var toMapMethod reflect.Value

				if copier.mapSuffix != "" {
					mapMethod := name + copier.mapSuffix
					if canAddr {
						toMapMethod = dest.Addr().MethodByName(mapMethod)
					} else {
						toMapMethod = dest.MethodByName(mapMethod)
					}
				}

				// has map method
				if toMapMethod.IsValid() && toMapMethod.Type().NumIn() == 1 && fromField.Type().AssignableTo(toMapMethod.Type().In(0)) {
					toMapMethod.Call([]reflect.Value{fromField})
				} else if toField := dest.FieldByName(name); toField.IsValid() {
					// has field
					if toField.CanSet() {
						if !set(toField, fromField) {
							if err := New(fromField.Interface(), toField.Addr().Interface(), copier.mapSuffix).Copy(); err != nil {
								return err
							}
						}
					}
				} else {
					// try to set to method
					var toMethod reflect.Value
					if canAddr {
						toMethod = dest.Addr().MethodByName(name)
					} else {
						toMethod = dest.MethodByName(name)
					}

					if toMethod.IsValid() && toMethod.Type().NumIn() == 1 && fromField.Type().AssignableTo(toMethod.Type().In(0)) {
						toMethod.Call([]reflect.Value{fromField})
					}
				}
			}
		}

		// Copy from method to field
		for _, field := range deepFields(toType) {
			name := field.Name
			var fromMethod reflect.Value
			if source.CanAddr() {
				fromMethod = source.Addr().MethodByName(name)
			} else {
				fromMethod = source.MethodByName(name)
			}

			if fromMethod.IsValid() && fromMethod.Type().NumIn() == 0 && fromMethod.Type().NumOut() == 1 {
				if toField := dest.FieldByName(field.Name); toField.IsValid() && toField.CanSet() {
					values := fromMethod.Call([]reflect.Value{})
					if len(values) >= 1 {
						set(toField, values[0])
					}
				}
			}
		}

		if copier.isSlice {
			if dest.Addr().Type().AssignableTo(copier.to.Type().Elem()) {
				copier.to.Set(reflect.Append(copier.to, dest.Addr()))
			} else if dest.Type().AssignableTo(copier.to.Type().Elem()) {
				copier.to.Set(reflect.Append(copier.to, dest))
			}
		}
	}
	return
}

func deepFields(reflectType reflect.Type) []reflect.StructField {
	var fields []reflect.StructField

	if reflectType = indirectType(reflectType); reflectType.Kind() == reflect.Struct {
		for i := 0; i < reflectType.NumField(); i++ {
			v := reflectType.Field(i)
			if v.Anonymous {
				fields = append(fields, deepFields(v.Type)...)
			} else {
				fields = append(fields, v)
			}
		}
	}

	return fields
}

func indirect(reflectValue reflect.Value) reflect.Value {
	for reflectValue.Kind() == reflect.Ptr {
		reflectValue = reflectValue.Elem()
	}
	return reflectValue
}

func indirectType(reflectType reflect.Type) reflect.Type {
	for reflectType.Kind() == reflect.Ptr || reflectType.Kind() == reflect.Slice {
		reflectType = reflectType.Elem()
	}
	return reflectType
}

func set(to, from reflect.Value) bool {
	if from.IsValid() {
		if to.Kind() == reflect.Ptr {
			//set `to` to nil if from is nil
			if from.Kind() == reflect.Ptr && from.IsNil() {
				to.Set(reflect.Zero(to.Type()))
				return true
			} else if to.IsNil() {
				to.Set(reflect.New(to.Type().Elem()))
			}
			to = to.Elem()
		}

		if from.Type().ConvertibleTo(to.Type()) {
			to.Set(from.Convert(to.Type()))
		} else if scanner, ok := to.Addr().Interface().(sql.Scanner); ok {
			err := scanner.Scan(from.Interface())
			if err != nil {
				return false
			}
		} else if from.Kind() == reflect.Ptr {
			return set(to, from.Elem())
		} else {
			return false
		}
	}
	return true
}
