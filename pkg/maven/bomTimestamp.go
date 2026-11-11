package maven

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"
)

const bomNamespace = "http://cyclonedx.org/schema/bom/1.4"

type bomXMLEdit struct {
	start, end int
	value      string
}

type bomXMLNode struct {
	name       xml.Name
	start      int
	attributes []xml.Attr
	text       strings.Builder
}

// EnsureBOMTimestamp repairs CycloneDX XML 1.4 metadata without round-tripping
// through a reduced BOM model. Unmodified documents are returned byte-for-byte.
// changed is false on failure; callers must not write output when err is non-nil.
// This validates XML structure and timestamp format, not full XSD compliance.
func EnsureBOMTimestamp(input []byte, now time.Time) (output []byte, changed bool, err error) {
	decoder := xml.NewDecoder(bytes.NewReader(input))
	var stack []*bomXMLNode
	var edits []bomXMLEdit
	metadataCount, timestampCount, rootCount := 0, 0, 0
	metadataEnd := 0
	metadataPrefix, emptyMetadataName := "", ""
	emptyMetadataStart := 0
	signed := false
	for {
		start := int(decoder.InputOffset())
		token, tokenErr := decoder.Token()
		if tokenErr == io.EOF {
			break
		}
		if tokenErr != nil {
			return nil, false, fmt.Errorf("invalid BOM XML: %w", tokenErr)
		}
		end := int(decoder.InputOffset())
		switch t := token.(type) {
		case xml.Directive:
			return nil, false, fmt.Errorf("XML directives/DTDs are not supported")
		case xml.StartElement:
			if len(stack) == 3 && stack[1].name == (xml.Name{Space: bomNamespace, Local: "metadata"}) && stack[2].name == (xml.Name{Space: bomNamespace, Local: "timestamp"}) {
				return nil, false, fmt.Errorf("metadata timestamp must contain text only")
			}
			if len(stack) == 0 {
				rootCount++
				if rootCount != 1 || t.Name.Local != "bom" || t.Name.Space != bomNamespace {
					return nil, false, fmt.Errorf("expected one CycloneDX XML 1.4 bom root")
				}
			}
			if t.Name.Local == "Signature" && t.Name.Space == "http://www.w3.org/2000/09/xmldsig#" {
				signed = true
			}
			if len(stack) == 1 && t.Name == (xml.Name{Space: bomNamespace, Local: "metadata"}) {
				metadataCount++
				metadataEnd = end
				raw := strings.TrimSpace(string(input[start:end]))
				name := strings.FieldsFunc(raw[1:], func(r rune) bool { return r == ' ' || r == '\t' || r == '\r' || r == '\n' || r == '/' || r == '>' })[0]
				if i := strings.IndexByte(name, ':'); i >= 0 {
					metadataPrefix = name[:i+1]
				}
				if strings.HasSuffix(raw, "/>") {
					emptyMetadataStart, emptyMetadataName = start, name
				}
			}
			stack = append(stack, &bomXMLNode{name: t.Name, start: start, attributes: t.Attr})
		case xml.CharData:
			if len(stack) > 0 {
				stack[len(stack)-1].text.Write([]byte(t))
			} else if strings.TrimSpace(string(t)) != "" {
				return nil, false, fmt.Errorf("text outside BOM root")
			}
		case xml.EndElement:
			node := stack[len(stack)-1]
			if len(stack) == 3 && stack[1].name == (xml.Name{Space: bomNamespace, Local: "metadata"}) && node.name == (xml.Name{Space: bomNamespace, Local: "timestamp"}) {
				timestampCount++
				if _, parseErr := time.Parse(time.RFC3339Nano, strings.TrimSpace(node.text.String())); parseErr != nil {
					return nil, false, fmt.Errorf("invalid metadata timestamp: %w", parseErr)
				}
			}
			if len(stack) == 4 && stack[1].name == (xml.Name{Space: bomNamespace, Local: "metadata"}) && stack[2].name == (xml.Name{Space: bomNamespace, Local: "properties"}) && node.name == (xml.Name{Space: bomNamespace, Local: "property"}) {
				for _, attr := range node.attributes {
					if attr.Name == (xml.Name{Local: "name"}) && attr.Value == "cdx:reproducible" {
						edits = append(edits, bomXMLEdit{start: node.start, end: end})
					}
				}
			}
			stack = stack[:len(stack)-1]
		}
	}
	if rootCount != 1 || metadataCount != 1 || timestampCount > 1 {
		return nil, false, fmt.Errorf("expected one BOM root, one metadata and at most one metadata timestamp")
	}
	if timestampCount == 1 {
		return input, false, nil
	}
	if signed {
		return nil, false, fmt.Errorf("refusing to modify a signed BOM")
	}
	timestamp := "<" + metadataPrefix + "timestamp>" + now.UTC().Format(time.RFC3339) + "</" + metadataPrefix + "timestamp>"
	if emptyMetadataName != "" {
		// Preserve namespace declarations/attributes on an empty metadata element.
		open := strings.TrimSuffix(string(input[emptyMetadataStart:metadataEnd]), "/>") + ">"
		edits = append(edits, bomXMLEdit{emptyMetadataStart, metadataEnd, open + timestamp + "</" + emptyMetadataName + ">"})
	} else {
		edits = append(edits, bomXMLEdit{metadataEnd, metadataEnd, timestamp})
	}
	sort.Slice(edits, func(i, j int) bool { return edits[i].start > edits[j].start })
	output = bytes.Clone(input)
	for _, edit := range edits {
		output = append(append(append([]byte{}, output[:edit.start]...), edit.value...), output[edit.end:]...)
	}
	// Validate the constructed output before the caller writes it.
	if _, _, err := EnsureBOMTimestamp(output, now); err != nil {
		return nil, false, fmt.Errorf("invalid post-processed BOM: %w", err)
	}
	return output, true, nil
}
