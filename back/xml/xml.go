package xml

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"time"
)

func Anonymize(xmlData []byte) ([]byte, error) {
	var buffer bytes.Buffer
	decoder := xml.NewDecoder(bytes.NewReader(xmlData))
	encoder := xml.NewEncoder(&buffer)

	var inFamily, inPatientPatient bool

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error decoding token: %w", err)
		}

		switch tok := token.(type) {
		case xml.StartElement:
			inFamily, inPatientPatient = handleStartElement(tok, encoder, inFamily, inPatientPatient)
		case xml.EndElement:
			inFamily, inPatientPatient = handleEndElement(tok, encoder, inFamily, inPatientPatient)
		case xml.CharData:
			handleCharData(tok, encoder, inFamily)
		default:
			encoder.EncodeToken(tok)
		}
	}

	if err := encoder.Flush(); err != nil {
		return nil, fmt.Errorf("error flushing encoder: %w", err)
	}

	return buffer.Bytes(), nil
}

func handleStartElement(tok xml.StartElement, encoder *xml.Encoder, inFamily, inPatientPatient bool) (bool, bool) {
	switch tok.Name.Local {
	case "family":
		inFamily = true
	case "birthTime":
		tok.Attr = removeDayFromBirthDate(tok.Attr)
	case "patientPatient":
		inPatientPatient = true
	case "id":
		if inPatientPatient {
			tok.Attr = modifyAttribute(tok.Attr, "extension", "")
		}
	}

	tok.Name.Space = ""
	tok.Attr = removeNamespace(tok.Attr)
	encoder.EncodeToken(tok)

	return inFamily, inPatientPatient
}

func handleEndElement(tok xml.EndElement, encoder *xml.Encoder, inFamily, inPatientPatient bool) (bool, bool) {
	if tok.Name.Local == "family" {
		inFamily = false
	} else if tok.Name.Local == "patientPatient" {
		inPatientPatient = false
	}

	tok.Name.Space = ""
	encoder.EncodeToken(tok)

	return inFamily, inPatientPatient
}

func handleCharData(tok xml.CharData, encoder *xml.Encoder, inFamily bool) {
	if inFamily {
		tok = []byte("")
	}
	encoder.EncodeToken(tok)
}

func removeDayFromBirthDate(attrs []xml.Attr) []xml.Attr {
	var newAttrs []xml.Attr
	for _, attr := range attrs {
		if attr.Name.Local == "value" {
			birth, err := time.Parse("20060102150405", attr.Value)
			if err != nil {
				attr.Value = fmt.Sprintln("Error:", err)
			} else {
				attr.Value = fmt.Sprintf("%d/%d", birth.Year(), birth.Month())
			}
		}
		newAttrs = append(newAttrs, attr)
	}
	return newAttrs
}

func modifyAttribute(attrs []xml.Attr, attrName, newValue string) []xml.Attr {
	var newAttrs []xml.Attr
	for _, attr := range attrs {
		if attr.Name.Local == attrName {
			attr.Value = newValue
		}
		newAttrs = append(newAttrs, attr)
	}
	return newAttrs
}

func removeNamespace(attrs []xml.Attr) []xml.Attr {
	for i, attr := range attrs {
		if attr.Name.Local == "xmlns" {
			return append(attrs[:i], attrs[i+1:]...)
		}
	}
	return attrs
}
