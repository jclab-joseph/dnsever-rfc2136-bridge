package dnshandler

import (
	"github.com/jclab-joseph/dnsever-rfc2136-bridge/dnsever"
	"github.com/jclab-joseph/dnsever-rfc2136-bridge/model"
)

type domainZone struct {
	// Zone domain name ending with a dot
	Zone     string
	Upstream []string
	Tsig     map[string]*tsigAuth
	Client   *dnsever.Client
}

type tsigAuth struct {
	Zone *domainZone
	model.TsigAuth
}
