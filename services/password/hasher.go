package password

type PasswordHasher interface {
	Hash(password []byte, salt []byte) (*HashSalt, error)
	Compare(hash, salt, password []byte) error
}
