package logger

func String(key, value string) Field {
	return Field{key, value}
}

func Error(err error) Field {
	return Field{
		Key:   "error",
		Value: err,
	}
}
