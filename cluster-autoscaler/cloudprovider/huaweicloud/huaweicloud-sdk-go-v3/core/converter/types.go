package converter

import "strconv"

type Int32Converter struct{}

func (i Int32Converter) CovertStringToInterface(value string) (interface{}, error) {
	i64, err := strconv.ParseInt(value, 10, 32)
	if err == nil {
		return int32(i64), nil
	}
	return int32(0), err
}

type Int64Converter struct{}

func (i Int64Converter) CovertStringToInterface(value string) (interface{}, error) {
	i64, err := strconv.ParseInt(value, 10, 64)
	if err == nil {
		return i64, nil
	}
	return int64(0), err
}

type Float32Converter struct{}

func (i Float32Converter) CovertStringToInterface(value string) (interface{}, error) {
	f64, err := strconv.ParseFloat(value, 32)
	if err == nil {
		return float32(f64), nil
	}
	return float32(0), err
}

type Float64Converter struct{}

func (i Float64Converter) CovertStringToInterface(value string) (interface{}, error) {
	f64, err := strconv.ParseFloat(value, 32)
	if err == nil {
		return f64, nil
	}
	return float64(0), err
}

type BooleanConverter struct{}

func (i BooleanConverter) CovertStringToInterface(value string) (interface{}, error) {
	boolean, err := strconv.ParseBool(value)
	if err == nil {
		return boolean, nil
	}
	return false, err
}

type StringConverter struct{}

func (i StringConverter) CovertStringToInterface(value string) (interface{}, error) {
	return value, nil
}
