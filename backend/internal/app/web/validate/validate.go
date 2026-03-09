package validate

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// 自定义校验器
func InitValidator() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("is_email", func(fl validator.FieldLevel) bool {
			s, ok := fl.Field().Interface().(string)
			if !ok {
				return false
			}
			// 自定义校验规则
			if len(s) < 5 {
				return false
			}
			return true
		})
	}
}

// 整理报错信息
func fieldLabel(fe validator.FieldError, obj any) string {
	// 通过反射拿 struct 字段上的 label 标签
	t := reflect.TypeOf(obj)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return fe.Field() // 兜底
	}
	if f, ok := t.FieldByName(fe.Field()); ok {
		if label := f.Tag.Get("label"); label != "" {
			return label
		}
	}
	return fe.Field()
}

// 把 validator 的错误翻译成人话
func translateValidationError(err error, obj any) string {
	var verrs validator.ValidationErrors
	if !errors.As(err, &verrs) {
		// 不是校验错误，直接返回原始信息或默认提示
		return "请求参数错误"
	}

	fe := verrs[0] // 先拿第一条就行，剩下的有需要可以再扩展
	field := fieldLabel(fe, obj)

	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s不能为空", field)
	case "min":
		return fmt.Sprintf("%s不能小于 %s", field, fe.Param())
	case "max":
		return fmt.Sprintf("%s不能大于 %s", field, fe.Param())
	case "oneof":
		return fmt.Sprintf("%s必须是以下值之一: %s", field, fe.Param())
	// 自定义 tag，例如 mobile
	case "mobile":
		return fmt.Sprintf("%s格式不正确", field)
	default:
		return fmt.Sprintf("%s不合法", field)
	}
}
