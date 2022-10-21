package cloudclient

type Client interface {
	CreateServer(spec *ServerSpec) (Server, error)
	GetServer(id string) (Server, error)
	DeleteServer(id string) error
}