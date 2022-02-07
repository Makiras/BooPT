package model

func UpdateIfNotEmpty(newValue, oldValue interface{}) interface{} {
	switch oldValue.(type) {
	case string:
		if newValue.(string) != "" {
			return newValue
		}
	case int:
		if newValue.(int) != 0 {
			return newValue
		}
	case int32:
		if newValue.(int32) != 0 {
			return newValue
		}
	case int64:
		if newValue.(int64) != 0 {
			return newValue
		}
	case float32:
		if newValue.(float32) != 0 {
			return newValue
		}
	case float64:
		if newValue.(float64) != 0 {
			return newValue
		}
	case bool:
		if newValue.(bool) {
			return newValue
		}
	case []byte:
		if len(newValue.([]byte)) != 0 {
			return newValue
		}
	case []string:
		if len(newValue.([]string)) != 0 {
			return newValue
		}
	case []int:
		if len(newValue.([]int)) != 0 {
			return newValue
		}
	case []int32:
		if len(newValue.([]int32)) != 0 {
			return newValue
		}
	case []int64:
		if len(newValue.([]int64)) != 0 {
			return newValue
		}
	case []float32:
		if len(newValue.([]float32)) != 0 {
			return newValue
		}
	case []float64:
		if len(newValue.([]float64)) != 0 {
			return newValue
		}
	case []bool:
		if len(newValue.([]bool)) != 0 {
			return newValue
		}
	case []*int64:
		if len(newValue.([]*int64)) != 0 {
			return newValue
		}
	}
	return nil
}
