package main

func isLower(c byte) bool {
	return c >= 'a' && c <= 'z'
}

func isUpper(c byte) bool {
	return c >= 'A' && c <= 'Z'
}

func isUnderscore(c byte) bool {
	return c == '_'
}

func isNumeric(c byte) bool {
	return c >= '0' && c <= '9'
}

func isAlphaNumeric(c byte) bool {
	return isLower(c) || isUpper(c) || isNumeric(c)
}

func isSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n'
}