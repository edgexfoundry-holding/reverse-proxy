package main

import (
	"fmt"
	"github.com/BurntSushi/toml"	
)

type tomlConfig struct {
	Title   	string
	KongUrl		kongurl
	KongAdmin   kongadmin
	SSLCertsPath	kongsslcerts
	EdgexServices	map[string]service
}

type kongurl struct{
	Server 		string
	AdminPort	string
	ApplicationPort string
}

type kongadmin struct{
	UserName string
	Password string
}

type kongsslcerts struct{
	CertPath string
	KeyPath	 string
	SNIS	 string
}

type service struct{
	Name		string
	Host		string
	Port		string
	Protocol 	string	
}


func LoadTomlConfig(path string) *tomlConfig {
	config :=tomlConfig{}
	if _, err := toml.DecodeFile(path, &config); err != nil {
		fmt.Println(err)
	}
	return &config
}