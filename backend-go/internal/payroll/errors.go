package payroll

import "errors"

var (
	// ErrRegionNotFound 地区未找到
	ErrRegionNotFound = errors.New("region not found")

	// ErrRuleNotFound 规则未找到
	ErrRuleNotFound = errors.New("region rule not found")

	// ErrTaxBracketsNotFound 税率表未找到
	ErrTaxBracketsNotFound = errors.New("tax brackets not found for year")

	// ErrInvalidGrossMonthly 无效的月薪
	ErrInvalidGrossMonthly = errors.New("grossMonthly must be greater than 0")

	// ErrCustomBaseRequired CUSTOM模式需要自定义基数
	ErrCustomBaseRequired = errors.New("custom base is required when mode is CUSTOM")

	// ErrCustomBaseMustBePositive 自定义基数必须大于0
	ErrCustomBaseMustBePositive = errors.New("custom base must be greater than 0")

	// ErrHFRateOutOfRange 公积金比例超出范围
	ErrHFRateOutOfRange = errors.New("housing fund rate is out of allowed range")

	// ErrSSRateOutOfRange 社保比例超出范围
	ErrSSRateOutOfRange = errors.New("social security rate is out of allowed range")
)

// ErrorCode 错误代码
type ErrorCode string

const (
	ErrorCodeInvalidArgument ErrorCode = "INVALID_ARGUMENT"
	ErrorCodeNotFound        ErrorCode = "NOT_FOUND"
	ErrorCodeInternal        ErrorCode = "INTERNAL_ERROR"
)
