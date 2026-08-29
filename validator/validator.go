package validator

import (
	"errors"
	"reflect"
	"sync"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/zh"
	"github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	zhtranslations "github.com/go-playground/validator/v10/translations/zh"
)

// DefaultValidator 验证器
type DefaultValidator struct {
	once     sync.Once
	validate *validator.Validate
	Trans    ut.Translator
	uni      *ut.UniversalTranslator
}

// ErrorValidator 自定义验证错误结构体
type ErrorValidator struct {
	Code       int    // 错误码
	Message    string // 错误消息
	StatusCode int    // 响应状态码
}

// Error 实现 error 接口
func (r *ErrorValidator) Error() string {
	return r.Message
}

var _ binding.StructValidator = &DefaultValidator{}

// ValidateStruct 验证结构体
// 如果接收到的类型是一个结构体或指向结构体的指针，则执行验证
func (v *DefaultValidator) ValidateStruct(obj any) error {
	if kindOfData(obj) == reflect.Struct {

		v.lazyinit()

		//如果传递不合规则的值，则返回InvalidValidationError，否则返回nil。
		///如果返回err != nil，可通过err.(validator.ValidationErrors)来访问错误数组。
		if err := v.validate.Struct(obj); err != nil {
			// 反回中文的第一条错误
			if errs, ok := err.(validator.ValidationErrors); ok {
				// return errors.New(errs[0].Translate(v.Trans))
				res := &ErrorValidator{
					Code:       400,
					Message:    errors.New(errs[0].Translate(v.Trans)).Error(),
					StatusCode: 200,
				}
				return res
			}
			return err
		}
	}
	return nil
}

// Engine 返回支持 StructValidator 实现的底层验证引擎
func (v *DefaultValidator) Engine() any {
	v.lazyinit()
	return v.validate
}

func (v *DefaultValidator) lazyinit() {
	v.once.Do(func() {
		v.validate = validator.New()
		v.validate.SetTagName("binding")
		zhCn := zh.New()
		v.uni = ut.New(zhCn, zhCn)
		v.Trans, _ = v.uni.GetTranslator("zh")

		// 注册一个函数，获取struct tag里自定义的label作为字段名
		_ = zhtranslations.RegisterDefaultTranslations(v.validate, v.Trans)

		v.validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := fld.Tag.Get("label")
			return name
		})
		v.validate.RegisterValidation("ipv4", IsIp4)
		v.validate.RegisterValidation("YYYY-MM-DD", IsYMD)
		v.validate.RegisterValidation("YYYY-MM-DD HH:mm", IsYMDHM)
		v.validate.RegisterValidation("YYYY-MM-DD HH:mm:ss", IsYMDHMS)
	})
}

func kindOfData(data any) reflect.Kind {
	value := reflect.ValueOf(data)
	valueType := value.Kind()

	if valueType == reflect.Ptr {
		valueType = value.Elem().Kind()
	}
	return valueType
}
