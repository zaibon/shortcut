//go:build dev

package env

func IsDev() bool {
	return true
}

func IsProd() bool {
	return false
}

func Name() string {
	return "dev"
}
