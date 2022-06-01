package auth

var key string

func SetKey(k string) {
	key = k
}

func IsAuthorized(k string) bool {
	return k == key
}
