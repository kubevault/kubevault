package secret

type SecretGetter interface {
	GetSecret() ([]byte, error)
}
