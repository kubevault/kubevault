package role

type RoleInterface interface {
	// CreateConfig creates database configuration
	CreateConfig() error

	// CreateRole creates role
	CreateRole() error
}
