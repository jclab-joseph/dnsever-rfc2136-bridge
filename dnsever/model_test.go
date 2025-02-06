package dnsever

import (
	"testing"
)

func TestUnmarshalDNSEver(t *testing.T) {
	xmlData := `<dnsever>
        <result type="gethost" code="700" numOfHosts="132" msg="Login Success" lang="kr">
                <host name="txt01.example.com" id="12345601" type="TXT" value="abcdef" zone="example.com" host="txt01"></host>
                <host name="txt02.example.com" id="12345602" type="TXT" value="test" zone="example.com" host="txt02"></host>
        </result>
    </dnsever>`

	d, err := UnmarshalDNSEver([]byte(xmlData))
	if err != nil {
		t.Fatalf("Failed to unmarshal XML: %v", err)
	}

	// Check result attributes
	if d.Result.Type != "gethost" {
		t.Errorf("Expected type 'gethost', got '%s'", d.Result.Type)
	}
	if d.Result.Code != 700 {
		t.Errorf("Expected code 700, got %d", d.Result.Code)
	}
	if d.Result.NumOfHosts != 132 {
		t.Errorf("Expected numOfHosts 132, got %d", d.Result.NumOfHosts)
	}
	if d.Result.Msg != "Login Success" {
		t.Errorf("Expected msg 'Login Success', got '%s'", d.Result.Msg)
	}
	if d.Result.Lang != "kr" {
		t.Errorf("Expected lang 'kr', got '%s'", d.Result.Lang)
	}

	// Check number of hosts
	if len(d.Result.Hosts) != 2 {
		t.Fatalf("Expected 2 hosts, got %d", len(d.Result.Hosts))
	}

	// Check first host
	host := d.Result.Hosts[0]
	if host.Name != "txt01.example.com" {
		t.Errorf("Expected name 'txt01.example.com', got '%s'", host.Name)
	}
	if host.ID != "12345601" {
		t.Errorf("Expected ID 12345601, got %s", host.ID)
	}
	if host.Type != "TXT" {
		t.Errorf("Expected type 'TXT', got '%s'", host.Type)
	}
	if host.Zone != "example.com" {
		t.Errorf("Expected zone 'example.com', got '%s'", host.Zone)
	}
	if host.Host != "txt01" {
		t.Errorf("Expected host 'txt01', got '%s'", host.Host)
	}
}
