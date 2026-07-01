package main

import "encoding/xml"

// Items represents the root element of a Burp Suite XML export.
type Items struct {
	XMLName     xml.Name `xml:"items"`
	BurpVersion string   `xml:"burpVersion,attr"`
	ExportTime  string   `xml:"exportTime,attr"`
	Items       []Item   `xml:"item"`
}

// Item represents a single request/response pair in the Burp export.
type Item struct {
	Time           string `xml:"time"`
	URL            string `xml:"url"`
	Host           Host   `xml:"host"`
	Port           int    `xml:"port"`
	Protocol       string `xml:"protocol"`
	Method         string `xml:"method"`
	Path           string `xml:"path"`
	Extension      string `xml:"extension"`
	Request        Data   `xml:"request"`
	Status         int    `xml:"status"`
	ResponseLength int    `xml:"responselength"`
	MimeType       string `xml:"mimetype"`
	Response       Data   `xml:"response"`
	Comment        string `xml:"comment"`
}

// Host represents the host element with an IP attribute.
type Host struct {
	IP    string `xml:"ip,attr"`
	Value string `xml:",chardata"`
}

// Data represents the request or response content, which may be base64 encoded.
type Data struct {
	Base64 bool   `xml:"base64,attr"`
	Value  string `xml:",chardata"`
}
