package engine

type EngineInterface interface {
	CreatePolicy() error
	UpdateAuthRole() error
	IsSecretEngineEnabled() (bool, error)
	EnableSecretEngine() error
	CreateConfig() error
}
