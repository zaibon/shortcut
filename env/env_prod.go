//go:build !dev

package env

func IsDev() bool {
	return false
}

func IsProd() bool {
	return true
}

func Name() string {
	return "prod"
}
