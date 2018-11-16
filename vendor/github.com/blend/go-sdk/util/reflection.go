package util

import (
	"encoding/base64"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/blend/go-sdk/exception"
)

var (
	// Reflection is a namespace for reflection utilities.
	Reflection = reflectionUtil{}
)

// PatchStrings options.
const (
	// FieldTagEnv is the struct tag for what environment variable to use to populate a field.
	FieldTagEnv = "env"
	// FieldFlagCSV is a field tag flag (say that 10 times fast).
	FieldFlagCSV = "csv"
	// FieldFlagBase64 is a field tag flag (say that 10 times fast).
	FieldFlagBase64 = "base64"
	// FieldFlagBytes is a field tag flag (say that 10 times fast).
	FieldFlagBytes = "bytes"
)

// Patcher describes an object that can be patched with raw values.
type Patcher interface {
	Patch(map[string]interface{}) error
}

// PatchStringer is a type that handles unmarshalling a map of strings into itself.
type PatchStringer interface {
	PatchStrings(map[string]string) error
}

type reflectionUtil struct{}

// FollowValuePointer derefs a reflectValue until it isn't a pointer, but will preseve it's nilness.
func (ru reflectionUtil) FollowValuePointer(v reflect.Value) interface{} {
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return nil
	}

	val := v
	for val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	return val.Interface()
}

// FollowType derefs a type until it isn't a pointer or an interface.
func (ru reflectionUtil) FollowType(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Ptr || t.Kind() == reflect.Interface {
		t = t.Elem()
	}
	return t
}

// FollowValue derefs a value until it isn't a pointer or an interface.
func (ru reflectionUtil) FollowValue(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	return v
}

// Value returns the integral reflect.Value for an object, derefing through pointers or interfaces.
func (ru reflectionUtil) Value(obj interface{}) reflect.Value {
	v := reflect.ValueOf(obj)
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	return v
}

// Type returns the integral type for an object, derefing through pointers or interfaces.
func (ru reflectionUtil) Type(obj interface{}) reflect.Type {
	t := reflect.TypeOf(obj)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	return t
}

// reflectSliceType returns the inner type of a slice following pointers.
func (ru reflectionUtil) SliceType(collection interface{}) reflect.Type {
	v := reflect.ValueOf(collection)
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Len() == 0 {
		t := v.Type()
		for t.Kind() == reflect.Ptr || t.Kind() == reflect.Slice {
			t = t.Elem()
		}
		return t
	}
	v = v.Index(0)
	for v.Kind() == reflect.Interface || v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	return v.Type()
}

// MakeNew returns a new instance of a reflect.Type.
func (ru reflectionUtil) MakeNew(t reflect.Type) interface{} {
	return reflect.New(t).Interface()
}

// MakeSliceOfType returns a new slice of a given reflect.Type.
func (ru reflectionUtil) MakeSliceOfType(t reflect.Type) interface{} {
	return reflect.New(reflect.SliceOf(t)).Interface()
}

// TypeName returns the string type name for an object's integral type.
func (ru reflectionUtil) TypeName(obj interface{}) string {
	return ru.Type(obj).Name()
}

// GetReflectValueByName returns a value for a given struct field by name.
func (ru reflectionUtil) GetValueByName(target interface{}, fieldName string) interface{} {
	targetValue := ru.Value(target)
	field := targetValue.FieldByName(fieldName)
	return field.Interface()
}

// GetReflectValueByName returns a value for a given struct field by name.
func (ru reflectionUtil) GetReflectFieldByName(target interface{}, fieldName string) interface{} {
	targetValue := ru.Value(target)
	return targetValue.FieldByName(fieldName)
}

func (ru reflectionUtil) getFieldByTag(objType reflect.Type, tagType string, tagName string) *reflect.StructField {
	for index := 0; index < objType.NumField(); index++ {
		field := objType.Field(index)

		tag := field.Tag
		fullTag := tag.Get(tagType)
		if String.CaseInsensitiveEquals(strings.Split(fullTag, ",")[0], tagName) {
			return &field
		}
	}
	return nil
}

func (ru reflectionUtil) SetValueByName(target interface{}, fieldName string, fieldValue interface{}) error {
	targetValue := ru.Value(target)
	targetType := ru.Type(target)
	return ru.SetValueByNameFromType(target, targetType, targetValue, fieldName, fieldValue)
}

// SetValueByNameFromTag sets a value on an object by its tag type/name combination
func (ru reflectionUtil) SetValueByNameFromTag(obj interface{}, objType reflect.Type, objValue reflect.Value, tagType string, tagName string, value interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = exception.New("panic setting value by tag").WithMessagef("tagType: %s tagName: % panic: %v", tagType, tagName, r)
		}
	}()

	relevantField := ru.getFieldByTag(objType, tagType, tagName)
	if relevantField == nil {
		err = exception.New("unknown tag").WithMessagef("%s %s `%s`", objType.Name(), tagType, tagName)
		return
	}

	return ru.doSetValue(*relevantField, objType, objValue, fmt.Sprintf("%s:%s", tagType, tagName), value)
}

// SetValueByNameFromType sets a value on an object by its field name.
func (ru reflectionUtil) SetValueByNameFromType(obj interface{}, objType reflect.Type, objValue reflect.Value, fieldName string, value interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = exception.New("panic setting value by name").WithMessagef("field: %s panic: %v", fieldName, r)
		}
	}()

	relevantField, hasField := objType.FieldByName(fieldName)
	if !hasField {
		err = exception.New("unknown field").WithMessagef("%s `%s`", objType.Name(), fieldName)
		return
	}

	return ru.doSetValue(relevantField, objType, objValue, fieldName, value)
}

func (ru reflectionUtil) doSetValue(relevantField reflect.StructField, objType reflect.Type, objValue reflect.Value, name string, value interface{}) (err error) {
	field := objValue.FieldByName(relevantField.Name)
	if !field.CanSet() {
		err = exception.New("cannot set field").WithMessagef("%s `%s`", objType.Name(), name)
		return
	}

	valueReflected := ru.Value(value)
	if !valueReflected.IsValid() {
		err = exception.New("invalid value").WithMessagef("%s `%s`", objType.Name(), name)
		return
	}

	assigned, assignErr := ru.tryAssignment(field, valueReflected)
	if assignErr != nil {
		err = assignErr
		return
	}
	if !assigned {
		err = exception.New("cannot set field").WithMessagef("%s `%s`", objType.Name(), name)
		return
	}
	return
}

// checks if a value is a zero value or its types default value
func (ru reflectionUtil) IsZero(v reflect.Value) bool {
	if !v.IsValid() {
		return true
	}
	switch v.Kind() {
	case reflect.Func, reflect.Map, reflect.Slice:
		return v.IsNil()
	case reflect.Array:
		z := true
		for i := 0; i < v.Len(); i++ {
			z = z && ru.IsZero(v.Index(i))
		}
		return z
	case reflect.Struct:
		z := true
		for i := 0; i < v.NumField(); i++ {
			z = z && ru.IsZero(v.Field(i))
		}
		return z
	}
	// Compare other types directly:
	z := reflect.Zero(v.Type())
	return v.Interface() == z.Interface()
}

// IsExported returns if a field is exported given its name and capitalization.
func (ru reflectionUtil) IsExported(fieldName string) bool {
	return fieldName != "" && strings.ToUpper(fieldName)[0] == fieldName[0]
}

// Decompose fully decomposes an object into map data.
func (ru reflectionUtil) Decompose(obj interface{}) map[string]interface{} {
	output := map[string]interface{}{}

	objMeta := ru.Type(obj)
	objValue := ru.Value(obj)

	var field reflect.StructField
	var fieldValue reflect.Value

	for x := 0; x < objMeta.NumField(); x++ {
		field = objMeta.Field(x)
		fieldValue = objValue.FieldByName(field.Name)

		if field.Type.Kind() == reflect.Struct || field.Type.Kind() == reflect.Slice || field.Type.Kind() == reflect.Map {
			output[field.Name] = ru.decomposeAny(objValue.Field(x).Interface())
		} else {
			output[field.Name] = fieldValue.Interface()
		}
	}

	return output
}

func (ru reflectionUtil) decomposeAny(obj interface{}) interface{} {
	objMeta := ru.Type(obj)
	objValue := ru.Value(obj)

	if objMeta.Kind() == reflect.Slice {
		output := make([]interface{}, objValue.Len())
		for index := 0; index < objValue.Len(); index++ {
			output[index] = ru.decomposeAny(objValue.Index(index).Interface())
		}
		return output
	}

	if objMeta.Kind() == reflect.Map {
		output := map[string]interface{}{}
		keys := objValue.MapKeys()
		for _, key := range keys {
			output[fmt.Sprintf("%v", key.Interface())] = ru.decomposeAny(objValue.MapIndex(key).Interface())
		}
	}

	if objMeta.Kind() == reflect.Struct {
		output := map[string]interface{}{}

		var field reflect.StructField
		var fieldValue reflect.Value

		for x := 0; x < objMeta.NumField(); x++ {
			field = objMeta.Field(x)
			fieldValue = objValue.FieldByName(field.Name)

			// Treat structs as nested values.
			if field.Type.Kind() == reflect.Struct || field.Type.Kind() == reflect.Slice || field.Type.Kind() == reflect.Map {
				output[field.Name] = ru.decomposeAny(objValue.Field(x).Interface())
			} else {
				output[field.Name] = fieldValue.Interface()
			}
		}
		return output
	}

	return obj
}

// DecomposeStrings decomposes an object into a string map.
func (ru reflectionUtil) DecomposeStrings(obj interface{}, tagName ...string) map[string]string {
	output := map[string]string{}

	objMeta := ru.Type(obj)
	objValue := ru.Value(obj)

	var field reflect.StructField
	var fieldValue reflect.Value
	var tag, tagValue string
	var dataField string
	var pieces []string
	var isCSV bool
	var isBytes bool
	var isBase64 bool

	if len(tagName) > 0 {
		tag = tagName[0]
	}

	for x := 0; x < objMeta.NumField(); x++ {
		isCSV = false
		isBytes = false
		isBase64 = false

		field = objMeta.Field(x)
		if !Reflection.IsExported(field.Name) {
			continue
		}

		fieldValue = objValue.FieldByName(field.Name)
		dataField = field.Name

		if field.Type.Kind() == reflect.Struct {
			childFields := ru.DecomposeStrings(fieldValue.Interface(), tagName...)
			for key, value := range childFields {
				output[key] = value
			}
		}

		if len(tag) > 0 {
			tagValue = field.Tag.Get(tag)
			if len(tagValue) > 0 {
				if field.Type.Kind() == reflect.Map {
					continue
				} else {
					pieces = strings.Split(tagValue, ",")
					dataField = pieces[0]

					if len(pieces) > 1 {
						for y := 1; y < len(pieces); y++ {
							if pieces[y] == FieldFlagCSV {
								isCSV = true
							} else if pieces[y] == FieldFlagBase64 {
								isBase64 = true
							} else if pieces[y] == FieldFlagBytes {
								isBytes = true
							}
						}
					}
				}
			}
		}

		if isCSV {
			if typed, isTyped := fieldValue.Interface().([]string); isTyped {
				output[dataField] = strings.Join(typed, ",")
			}
		} else if isBytes {
			if typed, isTyped := fieldValue.Interface().([]byte); isTyped {
				output[dataField] = string(typed)
			}
		} else if isBase64 {
			if typed, isTyped := fieldValue.Interface().([]byte); isTyped {
				output[dataField] = base64.StdEncoding.EncodeToString(typed)
			}
			if typed, isTyped := fieldValue.Interface().(string); isTyped {
				output[dataField] = typed
			}
		} else {
			output[dataField] = fmt.Sprintf("%v", ru.FollowValuePointer(fieldValue))
		}
	}

	return output
}

// Patch updates an object based on a map of field names to values.
func (ru reflectionUtil) Patch(obj interface{}, patchValues map[string]interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = exception.New(r)
		}
	}()

	if patchable, isPatchable := obj.(Patcher); isPatchable {
		return patchable.Patch(patchValues)
	}

	targetValue := ru.Value(obj)
	targetType := targetValue.Type()

	for key, value := range patchValues {
		err = ru.SetValueByNameFromType(obj, targetType, targetValue, key, value)
		if err != nil {
			return err
		}
	}
	return nil
}

// PatchByTag updates an object based on a map of tag names to values.
func (ru reflectionUtil) PatchByTag(obj interface{}, tagType string, patchValues map[string]interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = exception.New(r)
		}
	}()

	if patchable, isPatchable := obj.(Patcher); isPatchable {
		return patchable.Patch(patchValues)
	}

	targetValue := ru.Value(obj)
	targetType := targetValue.Type()

	for key, value := range patchValues {
		err = ru.SetValueByNameFromTag(obj, targetType, targetValue, tagType, key, value)
		if err != nil {
			return err
		}
	}
	return nil
}


// PatchStrings sets an object from a set of strings mapping field names to string values (to be parsed).
func (ru reflectionUtil) PatchStrings(tagName string, data map[string]string, obj interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = exception.New(r)
		}
	}()

	// check if the type implements marshaler.
	if typed, isTyped := obj.(PatchStringer); isTyped {
		return typed.PatchStrings(data)
	}

	objMeta := ru.Type(obj)
	objValue := ru.Value(obj)

	typeDuration := reflect.TypeOf(time.Duration(time.Nanosecond))

	var field reflect.StructField
	var fieldType reflect.Type
	var fieldValue reflect.Value
	var tag string
	var pieces []string
	var dataField string
	var dataValue string
	var dataFieldValue interface{}
	var hasDataValue bool

	var isCSV bool
	var isBytes bool
	var isBase64 bool
	var assigned bool

	for x := 0; x < objMeta.NumField(); x++ {
		isCSV = false
		isBytes = false
		isBase64 = false

		field = objMeta.Field(x)
		fieldValue = objValue.FieldByName(field.Name)

		// Treat structs as nested values.
		if field.Type.Kind() == reflect.Struct {
			if err = ru.PatchStrings(tagName, data, objValue.Field(x).Addr().Interface()); err != nil {
				return err
			}
			continue
		}

		tag = field.Tag.Get(tagName)
		if len(tag) > 0 {
			pieces = strings.Split(tag, ",")
			dataField = pieces[0]
			if len(pieces) > 1 {
				for y := 1; y < len(pieces); y++ {
					if pieces[y] == FieldFlagCSV {
						isCSV = true
					} else if pieces[y] == FieldFlagBase64 {
						isBase64 = true
					} else if pieces[y] == FieldFlagBytes {
						isBytes = true
					}
				}
			}

			dataValue, hasDataValue = data[dataField]
			if !hasDataValue {
				continue
			}

			if isCSV {
				dataFieldValue = strings.Split(dataValue, ",")
			} else if isBase64 {
				dataFieldValue, err = base64.StdEncoding.DecodeString(dataValue)
				if err != nil {
					return
				}
			} else if isBytes {
				dataFieldValue = []byte(dataValue)
			} else {
				// figure out the rootmost type (i.e. deref ****ptr etc.)
				fieldType = ru.FollowType(field.Type)
				switch fieldType {
				case typeDuration:
					dataFieldValue, err = time.ParseDuration(dataValue)
					if err != nil {
						err = exception.New(err)
						return
					}
				default:
					switch fieldType.Kind() {
					case reflect.Bool:
						if hasDataValue {
							dataFieldValue = Parse.Bool(dataValue)
						} else {
							continue
						}
					case reflect.Float32:
						dataFieldValue, err = strconv.ParseFloat(dataValue, 32)
						if err != nil {
							err = exception.New(err)
							return
						}
					case reflect.Float64:
						dataFieldValue, err = strconv.ParseFloat(dataValue, 64)
						if err != nil {
							err = exception.New(err)
							return
						}
					case reflect.Int8:
						dataFieldValue, err = strconv.ParseInt(dataValue, 10, 8)
						if err != nil {
							err = exception.New(err)
							return
						}
					case reflect.Int16:
						dataFieldValue, err = strconv.ParseInt(dataValue, 10, 16)
						if err != nil {
							return exception.New(err)
						}
					case reflect.Int32:
						dataFieldValue, err = strconv.ParseInt(dataValue, 10, 32)
						if err != nil {
							err = exception.New(err)
							return
						}
					case reflect.Int:
						dataFieldValue, err = strconv.ParseInt(dataValue, 10, 64)
						if err != nil {
							err = exception.New(err)
							return
						}
					case reflect.Int64:
						dataFieldValue, err = strconv.ParseInt(dataValue, 10, 64)
						if err != nil {
							return exception.New(err)
						}
					case reflect.Uint8:
						dataFieldValue, err = strconv.ParseUint(dataValue, 10, 8)
						if err != nil {
							err = exception.New(err)
							return
						}
					case reflect.Uint16:
						dataFieldValue, err = strconv.ParseUint(dataValue, 10, 8)
						if err != nil {
							err = exception.New(err)
							return
						}
					case reflect.Uint32:
						dataFieldValue, err = strconv.ParseUint(dataValue, 10, 32)
						if err != nil {
							err = exception.New(err)
							return
						}
					case reflect.Uint64:
						dataFieldValue, err = strconv.ParseUint(dataValue, 10, 64)
						if err != nil {
							err = exception.New(err)
							return
						}
					case reflect.Uint, reflect.Uintptr:
						dataFieldValue, err = strconv.ParseUint(dataValue, 10, 64)
						if err != nil {
							err = exception.New(err)
							return
						}
					case reflect.String:
						dataFieldValue = dataValue
					default:
						err = exception.New("map strings into; unhandled assignment").WithMessagef("type %s", fieldType.String())
						return
					}
				}
			}

			value := ru.Value(dataFieldValue)
			if !value.IsValid() {
				err = exception.New("invalid value").WithMessagef("%s `%s`", objMeta.Name(), field.Name)
				return
			}

			assigned, err = ru.tryAssignment(fieldValue, value)
			if err != nil {
				return
			}
			if !assigned {
				err = exception.New("cannot set field").WithMessagef("%s `%s`", objMeta.Name(), field.Name)
				return
			}
		}
	}
	return nil
}

func (ru reflectionUtil) tryAssignment(field, value reflect.Value) (assigned bool, err error) {
	if value.Type().AssignableTo(field.Type()) {
		field.Set(value)
		assigned = true
		return
	}

	if value.Type().ConvertibleTo(field.Type()) {
		convertedValue := value.Convert(field.Type())
		if convertedValue.Type().AssignableTo(field.Type()) {
			field.Set(convertedValue)
			assigned = true
			return
		}
	}

	if field.Type().Kind() == reflect.Ptr {
		if value.Type().AssignableTo(field.Type().Elem()) {
			elem := reflect.New(field.Type().Elem())
			elem.Elem().Set(value)
			field.Set(elem)
			assigned = true
			return
		} else if value.Type().ConvertibleTo(field.Type().Elem()) {
			elem := reflect.New(field.Type().Elem())
			elem.Elem().Set(value.Convert(field.Type().Elem()))
			field.Set(elem)
			assigned = true
			return
		}
	}

	return
}
