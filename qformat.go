package qformat

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type Q map[string]interface{}

var re = regexp.MustCompile("{.*?}")
var fX func() string
var fT = reflect.TypeOf(fX)

func (q *Q) Fprintf(writer io.Writer, s string, positionalArgs ...interface{}) {
	replacer := func(s string) string {
		stripped := s[1 : len(s)-1]
		if i, err := strconv.Atoi(stripped); err == nil {
			if i >= len(positionalArgs) {
				return fmt.Sprintf("<<Positional arg %d not present>>", i)
			} else {
				return fmt.Sprintf("%v", positionalArgs[i])
			}
		}

		parts := strings.Split(stripped, ".")

		if replacement, inMap := (*q)[parts[0]]; inMap {
			return extract(replacement, parts[1:])
		}
		if dot, hasDot := (*q)["."]; hasDot {
			return extract(dot, parts)
		}

		return fmt.Sprintf("<<Missing: '%s'>>", stripped)

	}

	//This is a hack, but beats fighting regexp
	s = strings.Replace(s, "%%", "%%p", -1)
	s = strings.Replace(s, "{{", "%%o", -1)
	s = strings.Replace(s, "}}", "%%c", -1)

	result := re.ReplaceAllStringFunc(s, replacer)

	result = strings.Replace(result, "%%c", "}", -1)
	result = strings.Replace(result, "%%o", "{", -1)
	result = strings.Replace(result, "%%p", "%%", -1)

	writer.Write([]byte(result))

}

func (q *Q) Sprintf(s string, positionalArgs ...interface{}) string {
	buf := new(bytes.Buffer)
	q.Fprintf(buf, s, positionalArgs...)
	return buf.String()
}

var emptyReflectValue = reflect.Value{}

func extract(v interface{}, parts []string) string {
	switch v := v.(type) {
	case func() string:
		return v()
	case string:
		return v
	}

	reflectValue := reflect.ValueOf(v)
	reflectType := reflectValue.Type()
	indirectValue := reflect.Indirect(reflectValue)
	indirectType := indirectValue.Type()

	if len(parts) == 0 {
		return fmt.Sprintf("%v", v)
	}

	if indirectType.Kind() == reflect.Struct {
		if fieldValue := indirectValue.FieldByName(parts[0]); fieldValue != emptyReflectValue {
			return extract(fieldValue.Interface(), parts[1:])
		}
	}

	if method, methodFound := reflectType.MethodByName(parts[0]); methodFound {
		if method.Type.NumIn() == 1 && method.Type.NumOut() == 1 {
			valueMethod := reflectValue.MethodByName(parts[0])
			resultInterface := valueMethod.Call([]reflect.Value{})[0].Interface()

			if method.Type.Out(0).Kind() == reflect.Ptr && method.Type.Out(0).Elem().Kind() == reflect.Struct {
				return extract(resultInterface, parts[1:])
			} else if method.Type.Out(0).Kind() == reflect.Interface {
				return extract(resultInterface, parts[1:])
			}

			return fmt.Sprintf("%v", resultInterface)
		}
	}

	pointerType := reflect.PtrTo(reflectType)
	if method, methodFound := pointerType.MethodByName(parts[0]); methodFound {
		if method.Type.NumIn() == 1 && method.Type.NumOut() == 1 {
			pointerValue := reflect.New(reflectType)
			tmp := pointerValue.Elem()
			tmp.Set(reflectValue)

			valueMethod := pointerValue.MethodByName(parts[0])
			return fmt.Sprintf("%v", valueMethod.Call([]reflect.Value{})[0].Interface())
		}
	}

	return fmt.Sprintf("<<Failed to find %v on %T>>", parts, v)
}

func Reflect(v interface{}) Q {
	q := Q{}
	q["."] = v
	return q
}

func (q *Q) Copy() Q {
	r := Q{}
	for k, v := range *q {
		r[k] = v
	}
	return r
}
