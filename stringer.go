package requests

type stringer interface {
	String() string
}

type stringerFn func() string

func (s stringerFn) String() string {
	return s()
}

type constString string

func (c constString) String() string {
	return string(c)
}

func toStringer(v interface{}) stringer {
	if s, ok := v.(stringer); ok {
		return s
	} else if s, ok := v.(stringerFn); ok {
		return s
	} else if s, ok := v.(func() string); ok {
		return stringerFn(s)
	} else if s, ok := v.(string); ok {
		return constString(s)
	} else if s, ok := v.(*string); ok {
		return stringerFn(func() string {
			return *s
		})
	} else {
		return nil
	}
}
