package logger

func String(key, value string) Field {
	return Field{key, value}
}

func Int64(key string, value int64) Field {
	return Field{
		Key:   key,
		Value: value,
	}
}

func Error(err error) Field {
	return Field{
		Key:   "error",
		Value: err,
	}
}
