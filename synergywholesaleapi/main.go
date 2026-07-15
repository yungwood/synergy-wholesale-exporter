package synergywholesaleapi

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

// I had trouble interacting with the Synergy Wholesale API using some existing SOAP libraries
// so I opted to write an implementation using core go libraries instead. I have only included
// the minimum required detail to work with the API

// Define structs for creating requests
type apiSOAPEnvelope struct {
	XMLName  xml.Name    `xml:"Envelope"`
	Xmlns    string      `xml:"xmlns:SOAP-ENV,attr"`
	Xmlns1   string      `xml:"xmlns:ns1,attr"`
	Xmlns2   string      `xml:"xmlns:ns2,attr"`
	XmlnsXSI string      `xml:"xmlns:xsi,attr"`
	Body     apiSOAPBody `xml:"Body"`
}

type apiSOAPBody struct {
	Content interface{} `xml:",any"`
}

// ListDomainsRequest defines your simple struct
type ListDomainsRequest struct {
	APIKey     string
	ResellerID string
}

func (r ListDomainsRequest) MarshalXML(encoder *xml.Encoder, start xml.StartElement) error {
	// Start the parent element (listDomains)
	start.Name.Local = "ns1:listDomains"
	if err := encoder.EncodeToken(start); err != nil {
		slog.Error("Error creating SOAP request", "error", err)
		return err
	}

	// Start the param element with xsi:type attribute
	paramStart := xml.StartElement{
		Name: xml.Name{Local: "param"},
		Attr: []xml.Attr{
			{Name: xml.Name{Local: "xsi:type"}, Value: "ns2:Map"},
		},
	}
	if err := encoder.EncodeToken(paramStart); err != nil {
		slog.Error("Error creating SOAP request", "error", err)
		return err
	}

	// Marshal the key-value pairs as nested items
	items := []struct {
		Key   string `xml:"key"`
		Value string `xml:"value"`
	}{
		{Key: "apiKey", Value: r.APIKey},
		{Key: "resellerID", Value: r.ResellerID},
	}

	for _, item := range items {
		if err := encoder.EncodeElement(item, xml.StartElement{Name: xml.Name{Local: "item"}}); err != nil {
			return err
		}
	}

	// End the param element
	if err := encoder.EncodeToken(paramStart.End()); err != nil {
		slog.Error("Error creating SOAP request", "error", err)
		return err
	}

	// End the listDomains element
	if err := encoder.EncodeToken(start.End()); err != nil {
		slog.Error("Error creating SOAP request", "error", err)
		return err
	}

	return nil
}

type ListDomainsResponse struct {
	XMLName xml.Name   `xml:"listDomainsResponse"`
	Return  DomainList `xml:"return"`
}

type DomainList struct {
	SOAPResponseCommon
	DomainList []DomainInfo `xml:"domainList>item"`
}

// This struct only unmarshalls the values that are useful for exporter purposes
type DomainInfo struct {
	Status         string            `xml:"status"`
	ErrorMessage   string            `xml:"errorMessage"`
	DomainName     string            `xml:"domainName"`
	DomainStatus   string            `xml:"domain_status"`
	DomainCreated  string            `xml:"domain_created,omitempty"`
	DomainExpiry   string            `xml:"domain_expiry,omitempty"`
	CreatedDate    string            `xml:"createdDate,omitempty"`
	TransferStatus string            `xml:"transfer_status,omitempty"`
	AutoRenew      int               `xml:"autoRenew,omitempty"`
	NameServers    []string          `xml:"nameServers>item"`
	DNSSECKeys     []DomainDNSSECKey `xml:"DSData>item"`
}

func (domainInfo DomainInfo) GetDomainExpiryTimestamp() int64 {
	return dateStringToTimestamp(domainInfo.DomainExpiry)
}

func (domainInfo DomainInfo) GetDomainCreationTimestamp() int64 {
	return dateStringToTimestamp(domainInfo.DomainCreated)
}

func dateStringToTimestamp(dateString string) int64 {
	if dateString == "" {
		return 0
	}
	// the timezone is not specified in the api documentation
	// based on expiry times for .com it seems to be Australia/Brisbane +1000
	location, err := time.LoadLocation("Australia/Brisbane")
	if err != nil {
		slog.Error("Error loading location", "error", err, "location", "Australia/Brisbane")
		return 0
	}
	layout := "2006-01-02 15:04:05"
	t, err := time.ParseInLocation(layout, dateString, location)
	if err != nil {
		return 0
	}
	utcTime := t.UTC()
	return utcTime.Unix()
}

type DomainDNSSECKey struct {
	KeyTag     string `xml:"keyTag"`
	Algorithm  string `xml:"algorithm"`
	DigestType string `xml:"digestType"`
	Digest     string `xml:"digest"`
}

type SOAPResponseCommon struct {
	Status       string `xml:"status"`
	ErrorMessage string `xml:"errorMessage,omitempty"`
}

type soapFault struct {
	FaultCode   string `xml:"faultcode"`
	FaultString string `xml:"faultstring"`
}

func (fault soapFault) Error() string {
	if fault.FaultCode == "" {
		return fmt.Sprintf("soap fault: %s", fault.FaultString)
	}
	return fmt.Sprintf("soap fault: %s: %s", fault.FaultCode, fault.FaultString)
}

type soapFaultEnvelope struct {
	XMLName xml.Name `xml:"Envelope"`
	Body    struct {
		Fault *soapFault `xml:"Fault"`
	} `xml:"Body"`
}

func createSOAPRequest(request interface{}) ([]byte, error) {
	envelope := apiSOAPEnvelope{
		Xmlns:    "http://schemas.xmlsoap.org/soap/envelope/",
		Xmlns2:   "http://xml.apache.org/xml-soap",
		Xmlns1:   "http://api.synergywholesale.com",
		XmlnsXSI: "http://www.w3.org/2001/XMLSchema-instance",
		Body: apiSOAPBody{
			Content: request,
		},
	}

	xmlBytes, err := xml.MarshalIndent(envelope, "", "  ")
	if err != nil {
		return nil, err
	}

	// Add the XML declaration header manually
	xmlHeader := []byte(xml.Header)
	xmlRequest := append(xmlHeader, xmlBytes...)

	return xmlRequest, nil
}

func Send(request ListDomainsRequest) (ListDomainsResponse, error) {
	response, err := SendSOAPRequest(request)
	if err != nil {
		return ListDomainsResponse{}, err
	}

	// Unmarshal the response
	var responseObject ListDomainsResponse
	err2 := UnmarshalSOAPResponse(response, &responseObject)
	if err2 != nil {
		return ListDomainsResponse{}, err2
	}
	return responseObject, nil
}

func SendSOAPRequest(param ListDomainsRequest) ([]byte, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	soapRequest, err := createSOAPRequest(param)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, "https://api.synergywholesale.com", bytes.NewBuffer(soapRequest))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "text/xml; charset=utf-8")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Log the response status code
	slog.Debug("Request successful", "response_code", resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("synergy wholesale api returned status %d: %s", resp.StatusCode, truncateBody(body, 1024))
	}

	return body, nil
}

func UnmarshalSOAPResponse(data []byte, response interface{}) error {
	var faultEnvelope soapFaultEnvelope
	if err := xml.NewDecoder(bytes.NewReader(data)).Decode(&faultEnvelope); err != nil {
		return fmt.Errorf("failed to unmarshal SOAP response: %w", err)
	}
	if faultEnvelope.Body.Fault != nil {
		return *faultEnvelope.Body.Fault
	}

	envelope := apiSOAPEnvelope{
		Body: apiSOAPBody{
			Content: response,
		},
	}

	decoder := xml.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&envelope)
	if err != nil {
		return fmt.Errorf("failed to unmarshal SOAP response: %w", err)
	}
	return nil
}

func truncateBody(body []byte, limit int) string {
	if len(body) <= limit {
		return string(body)
	}
	return string(body[:limit]) + "...[truncated]"
}
