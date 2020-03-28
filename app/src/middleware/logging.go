package middleware

func SuspiciousRequestLoggerMiddleware(f Handler) Handler {
	//Use warn
	return f
}

func HTTPErrorLoggerMiddleware(f Handler) Handler {
	//use warn
	return f
}
