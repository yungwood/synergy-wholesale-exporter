package synergywholesaleapi

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestUnmarshalSOAPResponseListDomains(t *testing.T) {
	data, err := os.ReadFile("testdata/list_domains_response.xml")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	var response ListDomainsResponse
	if err := unmarshalSOAPResponse(data, &response); err != nil {
		t.Fatalf("unmarshal SOAP response: %v", err)
	}

	if response.Return.Status != "OK" {
		t.Fatalf("top-level status = %q, want OK", response.Return.Status)
	}

	domains := response.Return.DomainList
	if len(domains) != 2 {
		t.Fatalf("domain count = %d, want 2", len(domains))
	}

	domain := domains[0]
	if domain.Status != "OK" {
		t.Errorf("domain status = %q, want OK", domain.Status)
	}
	if domain.DomainName != "example.com" {
		t.Errorf("domain name = %q, want example.com", domain.DomainName)
	}
	if domain.DomainStatus != "ok" {
		t.Errorf("domain registry status = %q, want ok", domain.DomainStatus)
	}
	if domain.AutoRenew != 1 {
		t.Errorf("auto renew = %d, want 1", domain.AutoRenew)
	}
	if domain.DomainExpiry != "2025-01-02 03:04:05" {
		t.Errorf("domain expiry = %q, want 2025-01-02 03:04:05", domain.DomainExpiry)
	}
	if domain.GetDomainExpiryTimestamp() == 0 {
		t.Error("domain expiry timestamp = 0, want non-zero timestamp")
	}

	wantNameServers := []string{"ns1.example.net", "ns2.example.net"}
	if len(domain.NameServers) != len(wantNameServers) {
		t.Fatalf("nameserver count = %d, want %d", len(domain.NameServers), len(wantNameServers))
	}
	for i, want := range wantNameServers {
		if domain.NameServers[i] != want {
			t.Errorf("nameserver[%d] = %q, want %q", i, domain.NameServers[i], want)
		}
	}

	if len(domain.DNSSECKeys) != 1 {
		t.Fatalf("DNSSEC key count = %d, want 1", len(domain.DNSSECKeys))
	}
	dnsSECKey := domain.DNSSECKeys[0]
	if dnsSECKey.KeyTag != "12345" {
		t.Errorf("DNSSEC key tag = %q, want 12345", dnsSECKey.KeyTag)
	}
	if dnsSECKey.Algorithm != "13" {
		t.Errorf("DNSSEC algorithm = %q, want 13", dnsSECKey.Algorithm)
	}
	if dnsSECKey.DigestType != "2" {
		t.Errorf("DNSSEC digest type = %q, want 2", dnsSECKey.DigestType)
	}
	if dnsSECKey.Digest != "0000000000000000000000000000000000000000000000000000000000000001" {
		t.Errorf("DNSSEC digest = %q, want sanitized digest", dnsSECKey.Digest)
	}

	missingDomain := domains[1]
	if missingDomain.Status != "ERR_DOMAIN_NOT_FOUND" {
		t.Errorf("missing domain status = %q, want ERR_DOMAIN_NOT_FOUND", missingDomain.Status)
	}
	if missingDomain.ErrorMessage != "Domain name does not exist." {
		t.Errorf("missing domain error = %q, want Domain name does not exist.", missingDomain.ErrorMessage)
	}
	if missingDomain.DomainName != "missing.example" {
		t.Errorf("missing domain name = %q, want missing.example", missingDomain.DomainName)
	}
}

func TestUnmarshalSOAPResponseFault(t *testing.T) {
	data := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/">
  <SOAP-ENV:Body>
    <SOAP-ENV:Fault>
      <faultcode>SOAP-ENV:Client</faultcode>
      <faultstring>Invalid API key</faultstring>
    </SOAP-ENV:Fault>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`)

	var response ListDomainsResponse
	err := unmarshalSOAPResponse(data, &response)
	if err == nil {
		t.Fatal("unmarshal SOAP fault error = nil, want error")
	}
	if !strings.Contains(err.Error(), "SOAP-ENV:Client") {
		t.Errorf("fault error = %q, want fault code", err.Error())
	}
	if !strings.Contains(err.Error(), "Invalid API key") {
		t.Errorf("fault error = %q, want fault string", err.Error())
	}
}

func TestCreateSOAPRequestListDomains(t *testing.T) {
	data, err := createSOAPRequest(ListDomainsRequest{
		APIKey:     "test-api-key",
		ResellerID: "12345",
	})
	if err != nil {
		t.Fatalf("create SOAP request: %v", err)
	}

	request := string(data)
	for _, want := range []string{
		`<?xml version="1.0" encoding="UTF-8"?>`,
		`xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/"`,
		`xmlns:ns1="http://api.synergywholesale.com"`,
		`xmlns:ns2="http://xml.apache.org/xml-soap"`,
		`xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"`,
		`<ns1:listDomains>`,
		`<param xsi:type="ns2:Map">`,
		`<key>apiKey</key>`,
		`<value>test-api-key</value>`,
		`<key>resellerID</key>`,
		`<value>12345</value>`,
	} {
		if !strings.Contains(request, want) {
			t.Errorf("SOAP request missing %q:\n%s", want, request)
		}
	}
}

func TestDateStringToTimestamp(t *testing.T) {
	timestamp := dateStringToTimestamp("2025-01-02 03:04:05")
	if timestamp != 1735751045 {
		t.Errorf("timestamp = %d, want 1735751045", timestamp)
	}

	if timestamp := dateStringToTimestamp(""); timestamp != 0 {
		t.Errorf("empty timestamp = %d, want 0", timestamp)
	}

	if timestamp := dateStringToTimestamp("not-a-date"); timestamp != 0 {
		t.Errorf("invalid timestamp = %d, want 0", timestamp)
	}
}

func TestSendSOAPRequestReturnsErrorForNon2xxStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			t.Errorf("request method = %s, want POST", request.Method)
		}
		if got := request.Header.Get("Content-Type"); got != "text/xml; charset=utf-8" {
			t.Errorf("content type = %q, want text/xml; charset=utf-8", got)
		}

		http.Error(writer, "upstream unavailable", http.StatusServiceUnavailable)
	}))
	defer server.Close()

	originalAPIEndpoint := apiEndpoint
	apiEndpoint = server.URL
	t.Cleanup(func() {
		apiEndpoint = originalAPIEndpoint
	})

	_, err := sendSOAPRequest(ListDomainsRequest{
		APIKey:     "test-api-key",
		ResellerID: "12345",
	})
	if err == nil {
		t.Fatal("send SOAP request error = nil, want error")
	}
	if !strings.Contains(err.Error(), "status 503") {
		t.Errorf("error = %q, want status 503", err.Error())
	}
	if !strings.Contains(err.Error(), "upstream unavailable") {
		t.Errorf("error = %q, want response body snippet", err.Error())
	}
}
