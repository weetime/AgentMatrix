package kit

import (
	"fmt"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/golang/protobuf/ptypes/wrappers"
	"k8s.io/apimachinery/pkg/api/resource"
)

var validate = validator.New(validator.WithRequiredStructEnabled())

var (
	abspathRegex = regexp.MustCompile(`^/[a-zA-Z0-9/_.-]*$`)
	// 规则根据 https://github.com/opencontainers/distribution-spec/blob/main/spec.md#pulling-manifests
	imageNameRegex = regexp.MustCompile(`^[a-z0-9]+((\.|_|__|-+)[a-z0-9]+)*(\/[a-z0-9]+((\.|_|__|-+)[a-z0-9]+)*)*$`)
	// 规则根据 https://github.com/opencontainers/distribution-spec/blob/main/spec.md#pulling-manifests
	imageTagRegex    = regexp.MustCompile(`^[a-zA-Z0-9_][a-zA-Z0-9._-]{0,127}$`)
	splitParamsRegex = regexp.MustCompile(`'[^']*'|\S+`)
)

var (
	oneofValsCache       = map[string][]string{}
	oneofValsCacheRWLock = sync.RWMutex{}
)

func init() {
	// kubernetes name format
	validate.RegisterAlias("alias_name", "min=1,max=50")
	validate.RegisterAlias("k8s_name_strict", "hostname_rfc1123,min=1,max=253")
	validate.RegisterAlias("port", "min=1,max=65535")
	validate.RegisterValidation("abspath", func(fl validator.FieldLevel) bool {
		return abspathRegex.MatchString(fl.Field().String())
	})
	validate.RegisterValidation("domain", func(fl validator.FieldLevel) bool {
		parsedURL, err := url.Parse(fl.Field().String())
		if err != nil {
			return false
		}
		return parsedURL.Scheme != "" && parsedURL.Host != ""
	})
	// docker image format
	validate.RegisterValidation("image", func(fl validator.FieldLevel) bool {
		_, _, _, err := parseArtifact(fl.Field().String())
		return err == nil
	})
	// docker image's name format
	validate.RegisterValidation("image_name", func(fl validator.FieldLevel) bool {
		return imageNameRegex.MatchString(fl.Field().String())
	})
	// docker image's tag format
	validate.RegisterValidation("image_tag", func(fl validator.FieldLevel) bool {
		return imageTagRegex.MatchString(fl.Field().String())
	})
	// map_required=value Field 表示当前字段为value时, Field字段必填
	validate.RegisterValidation("map_required", func(fl validator.FieldLevel) bool {
		params := parseOneOfParam2(fl.Param())
		if len(params)%2 != 0 {
			panic(fmt.Sprintf("Bad param number for map_required %s", fl.FieldName()))
		}

		field := fl.Field()
		var v string
		switch field.Kind() {
		case reflect.String:
			v = field.String()
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			v = strconv.FormatInt(field.Int(), 10)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			v = strconv.FormatUint(field.Uint(), 10)
		default:
			panic(fmt.Sprintf("Bad field type %T", field.Interface()))
		}

		for i := 0; i < len(params); i += 2 {
			value, param := params[i], params[i+1]
			if v == value {
				field, _, _, found := fl.GetStructFieldOKAdvanced2(fl.Parent(), param)
				if !found || !hasValue(field) {
					return false
				}
			}
		}
		return true
	})
	// kubernetes size format
	validate.RegisterValidation("size", func(fl validator.FieldLevel) bool {
		_, err := resource.ParseQuantity(fl.Field().String())
		return err == nil
	})
	// validate wrappers.StringValue.Value instead of wrappers.StringValue
	validate.RegisterCustomTypeFunc(func(field reflect.Value) interface{} {
		return field.Interface().(wrappers.StringValue).Value
	}, wrappers.StringValue{})
	validate.RegisterCustomTypeFunc(func(field reflect.Value) interface{} {
		return field.Interface().(wrappers.BoolValue).Value
	}, wrappers.BoolValue{})
	validate.RegisterValidation("time", func(fl validator.FieldLevel) bool {
		return checkTimeStrInvalid(fl.Field().String())
	})
}

func Validate(v interface{}) error {
	if err := validate.Struct(v); err != nil {
		return err
	}
	return nil
}

func ValidateVar(v interface{}, tag string) error {
	return validate.Var(v, tag)
}

func ValidateOfTag(v interface{}) error {
	rv := reflect.Indirect(reflect.ValueOf(v))
	if rv.Kind() != reflect.Struct {
		return nil
	}

	for i := 0; i < rv.NumField(); i++ {
		field := rv.Type().Field(i)
		if field.Tag.Get("validate") == "required" && rv.Field(i).IsZero() {
			return fmt.Errorf("%s field %s is required", rv.Type().Name(), field.Name)
		}
	}
	return nil
}

// ValidateFieldNames 验证结构体中的指定字段是否为零值
func ValidateOfFieldNames(v interface{}, fieldNames ...string) error {
	val := reflect.ValueOf(v)
	var structVal reflect.Value

	// 检查是否为指针，并获取指针指向的值
	if val.Kind() == reflect.Ptr {
		structVal = val.Elem()
		if !structVal.IsValid() {
			return fmt.Errorf("pointer is nil")
		}
	} else if val.Kind() == reflect.Struct {
		structVal = val
	} else {
		return fmt.Errorf("provided value is not a struct or a pointer to a struct")
	}

	for _, fieldName := range fieldNames {
		fieldVal := structVal.FieldByName(fieldName)
		if !fieldVal.IsValid() {
			return fmt.Errorf("field does not exist")
		}

		zero := reflect.Zero(fieldVal.Type())
		if reflect.DeepEqual(fieldVal.Interface(), zero.Interface()) {
			return fmt.Errorf("%s field %s is required", val.Type().Name(), fieldName)
		}
	}

	return nil
}

func parseOneOfParam2(s string) []string {
	oneofValsCacheRWLock.RLock()
	vals, ok := oneofValsCache[s]
	oneofValsCacheRWLock.RUnlock()
	if !ok {
		oneofValsCacheRWLock.Lock()
		vals = splitParamsRegex.FindAllString(s, -1)
		for i := 0; i < len(vals); i++ {
			vals[i] = strings.Replace(vals[i], "'", "", -1)
		}
		oneofValsCache[s] = vals
		oneofValsCacheRWLock.Unlock()
	}
	return vals
}

func hasValue(field reflect.Value) bool {
	switch field.Kind() {
	case reflect.Slice, reflect.Map, reflect.Ptr, reflect.Interface, reflect.Chan, reflect.Func:
		return !field.IsNil()
	default:
		if field.CanInterface() && field.Interface() != nil {
			return true
		}
		return field.IsValid() && !field.IsZero()
	}
}

func checkTimeStrInvalid(timeStr string) bool {
	if timeStr == "" {
		return true
	}
	if _, err := time.Parse(time.DateTime, timeStr); err != nil {
		return false
	}
	return true
}
