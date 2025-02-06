package dnsever

import "encoding/xml"

type DNSEverXml struct {
	XMLName xml.Name  `xml:"dnsever"`
	Result  ResultXml `xml:"result"`
}

type ResultXml struct {
	Type       string    `xml:"type,attr"`
	Code       int       `xml:"code,attr"`
	NumOfHosts int       `xml:"numOfHosts,attr"`
	Msg        string    `xml:"msg,attr"`
	Lang       string    `xml:"lang,attr"`
	Hosts      []HostXml `xml:"host"`
}

type HostXml struct {
	Name  string `xml:"name,attr"`
	ID    string `xml:"id,attr"`
	Type  string `xml:"type,attr"`
	Value string `xml:"value,attr"`
	Zone  string `xml:"zone,attr"`
	Host  string `xml:"host,attr"`
}

func UnmarshalDNSEver(data []byte) (*DNSEverXml, error) {
	var d DNSEverXml
	err := xml.Unmarshal(data, &d)
	if err != nil {
		return nil, err
	}
	return &d, nil
}
