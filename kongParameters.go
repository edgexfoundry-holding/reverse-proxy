package main

type KongService struct {
	Name string `url:"name,omitempty"`
	Host string `url:"host,omitempty"`
	Port string	`url:"port,omitempty"`
	Protocol string `url:"protocol,omitempty"`
}

type KongRoute struct {
	Paths []string `url:"paths[],omitempty"`
}

type KongPlugin struct {
	Name string `url:"name,omitempty"`
}

type KongBasicAuthPlugin struct{
	Name string `url:"name,omitempty"`
	HideCredentials string `url:"config.hide_credentials,omitempty"`
}

type KongUser struct {
	UserName string `url:"username,omitempty"`
	Password string `url:"password,omitempty"`
}

type KongCert struct {
	Cert string `url:cerpath,omitempty`
	Key string  `url:keypath, omitempty`
	Snis string `url:snis,  omitempty`
}
