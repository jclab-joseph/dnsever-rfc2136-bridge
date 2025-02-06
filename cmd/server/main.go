package main

import (
	"flag"
	"github.com/jclab-joseph/dnsever-rfc2136-bridge/dnshandler"
	"github.com/jclab-joseph/dnsever-rfc2136-bridge/model"
	"gopkg.in/yaml.v3"
	"log"
	"os"

	"github.com/miekg/dns"
)

func getEnvOrDefault(name string, def string) string {
	value := os.Getenv(name)
	if len(value) <= 0 {
		value = def
	}
	return value
}

func main() {
	var authFilePath string
	var listen string
	flag.StringVar(&listen, "listen", getEnvOrDefault("LISTEN", ":2053"), "LISTEN\tAddress to listen on")
	flag.StringVar(&authFilePath, "auth-file", getEnvOrDefault("AUTH_FILE", ""), "AUTH_FILE\tauth yaml file path")
	flag.Parse()

	authFileRaw, err := os.ReadFile(authFilePath)
	if err != nil {
		log.Fatalf("Failed to read auth file(%s): %+v", authFilePath, err)
	}

	var authFile model.AuthFile
	if err := yaml.Unmarshal(authFileRaw, &authFile); err != nil {
		log.Fatalf("Failed to read auth file: %+v", err)
	}

	handler := &dnshandler.Handler{}
	handler.ApplyAuthFile(&authFile)

	server := &dns.Server{
		Addr:          listen,
		Net:           "tcp",
		TsigSecret:    make(map[string]string),
		Handler:       handler,
		MsgAcceptFunc: handler.MsgAccept,
	}
	for _, domain := range authFile.Domains {
		for _, auth := range domain.Tsig {
			server.TsigSecret[auth.Name] = auth.Secret
		}
	}

	log.Printf("Starting DNS UPDATE server on %s", listen)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Failed to start server: %s", err.Error())
	}
}
