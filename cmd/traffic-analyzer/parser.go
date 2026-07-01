package main

import (
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"os"
)

// ParseBurpXML reads a Burp Suite XML export and returns the unmarshaled Items.
func ParseBurpXML(filename string) (*Items, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	byteValue, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	var items Items
	if err := xml.Unmarshal(byteValue, &items); err != nil {
		return nil, fmt.Errorf("error unmarshaling XML: %w", err)
	}

	return &items, nil
}

// DecodeData decodes the base64 value of a Data struct if needed.
func DecodeData(data Data) ([]byte, error) {
	if !data.Base64 {
		return []byte(data.Value), nil
	}

	decoded, err := base64.StdEncoding.DecodeString(data.Value)
	if err != nil {
		return nil, fmt.Errorf("error decoding base64: %w", err)
	}

	return decoded, nil
}
