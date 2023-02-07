package _import

type CredHelper struct {
	registryUser     string
	registryPassword string
}

func NewCredHelper(registryUser string, registryPassword string) *CredHelper {
	return &CredHelper{
		registryUser:     registryUser,
		registryPassword: registryPassword,
	}
}

func (c *CredHelper) Get(serverURL string) (string, string, error) {
	return c.registryUser, c.registryPassword, nil
}
