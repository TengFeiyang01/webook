package logger

func String(key, value string) Field {
	return Field{key, value}
}
