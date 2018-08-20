package types

import (
	"io"
)

// CloneRequest is the structure to
// use when calling dahu-git for cloning
// a repository
type CloneRequest struct {
	SshAuth    SshAuth  `json:"sshAuth"`
	HttpAuth   HttpAuth `json:"httpAuth"`
	Branch     string   `json:"branch"`
	NoCheckout bool     `json:"noCheckout"`
	UseSsh     bool     `json:"useSsh"`
	UseHttp    bool     `json:"useHttp"`
}

// SshAuth contains all informations
// concerning the ssh authentication and connection
type SshAuth struct {
	Url         string `json:"url"`
	Key         string `json:"key"`
	KeyPassword string `json:"keyPassword"`
}

// HttpAuth contains all informations
// concerning the http authentication and connection
type HttpAuth struct {
	Url      string `json:"url"`
	User     string `json:"user"`
	Password string `json:"password"`
}

type CloneContext struct {
	Directory  string
	NoCheckout bool
	Branch     string
	Progress   io.Writer
}
