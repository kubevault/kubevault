package docker

const (
	ACRegistry = "soter"
	ImageStash = "vault-operator"
)

type Docker struct {
	Registry, Image, Tag string
}

func (docker Docker) ToContainerImage() string {
	return docker.Registry + "/" + docker.Image + ":" + docker.Tag
}
