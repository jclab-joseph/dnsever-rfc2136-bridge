package main

import (
	"flag"
	"github.com/jclab-joseph/dnsever-rfc2136-bridge/dnshandler"
	"github.com/jclab-joseph/dnsever-rfc2136-bridge/model"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"sync"

	"github.com/miekg/dns"
)

func getEnvOrDefault(name string, def string) string {
	value := os.Getenv(name)
	if len(value) <= 0 {
		value = def
	}
	return value
}

type App struct {
	handler   *dnshandler.Handler
	tcpServer *dns.Server
	udpServer *dns.Server
}

func (a *App) addTsigSecret(name string, secret string) {
	a.tcpServer.TsigSecret[name] = secret
	a.udpServer.TsigSecret[name] = secret
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

	app := &App{
		handler: handler,
		tcpServer: &dns.Server{
			Addr:          listen,
			Net:           "tcp",
			TsigSecret:    make(map[string]string),
			Handler:       handler,
			MsgAcceptFunc: handler.MsgAccept,
		},
		udpServer: &dns.Server{
			Addr:          listen,
			Net:           "udp",
			TsigSecret:    make(map[string]string),
			Handler:       handler,
			MsgAcceptFunc: handler.MsgAccept,
		},
	}
	for _, domain := range authFile.Domains {
		for _, auth := range domain.Tsig {
			app.addTsigSecret(auth.Name, auth.Secret)
		}
	}

	var wg sync.WaitGroup

	log.Printf("Starting DNS UPDATE server on %s", listen)

	wg.Add(2)
	go func() {
		defer wg.Done()
		if err := app.tcpServer.ListenAndServe(); err != nil {
			log.Fatalf("Failed to start tcp server: %s", err.Error())
		}
	}()
	go func() {
		defer wg.Done()
		if err := app.udpServer.ListenAndServe(); err != nil {
			log.Fatalf("Failed to start udp server: %s", err.Error())
		}
	}()
	wg.Wait()
}
