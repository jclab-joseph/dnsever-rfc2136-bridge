package model

type TsigAuth struct {
	Name   string `json:"name" yaml:"name"`
	Secret string `json:"secret" yaml:"secret"`
}

type Domain struct {
	Zone         string     `json:"zone" yaml:"zone"`
	Upstream     []string   `json:"upstream" yaml:"upstream"`
	ClientId     string     `json:"clientId" yaml:"clientId"`
	ClientSecret string     `json:"clientSecret" yaml:"clientSecret"`
	Tsig         []TsigAuth `json:"tsig" yaml:"tsig"`
}

type AuthFile struct {
	Domains []Domain `json:"domains" yaml:"domains"`
}
