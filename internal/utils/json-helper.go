package utils

type Dictionary map[string]any

func (d Dictionary) UnwindValue(keys ...string) any {
	current := d
	for _, key := range keys {
		if current == nil {
			return nil
		}
		if val, ok := current[key]; ok {
			switch v := val.(type) {
			case map[string]any:
				current = Dictionary(v)
			case Dictionary:
				current = v
			default:
				return val
			}
		} else {
			return nil
		}
	}
	return current
}

func (d Dictionary) UnwindString(keys ...string) string {
	if val := d.UnwindValue(keys...); val != nil {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func (d Dictionary) UnwindBool(keys ...string) bool {
	if val := d.UnwindValue(keys...); val != nil {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}

func (d Dictionary) UnwindInt(keys ...string) int {
	if val := d.UnwindValue(keys...); val != nil {
		switch v := val.(type) {
		case int:
			return v
		case int32:
			return int(v)
		case int64:
			return int(v)
		case float64:
			return int(v)
		case float32:
			return int(v)
		}
	}
	return 0
}
