package cloudclient

var _ Client = &CloudClient{}

type CloudClient struct {
	Url string
}

type ServerSpec struct {
	Name string `json:"name,omitempty"`
}

type Server struct {
	Name string 
	Id string
}

type HttpError struct {
	Status int
	Message string
}