package maven

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestEnsureBOMTimestamp(t *testing.T) {
	now := time.Date(2026, 10, 22, 12, 34, 56, 123, time.FixedZone("test", 2*60*60))
	timestamp := "<timestamp>2026-10-22T10:34:56Z</timestamp>"
	wrap := func(metadata string) string {
		return `<bom xmlns="http://cyclonedx.org/schema/bom/1.4" serialNumber="urn:uuid:test" version="1">` + metadata + `</bom>`
	}
	for _, tc := range []struct {
		name, input, want string
	}{
		{"missing", wrap(`<metadata><component bom-ref="root"/></metadata>`), wrap(`<metadata>` + timestamp + `<component bom-ref="root"/></metadata>`)},
		{"empty metadata", wrap(`<metadata/>`), wrap(`<metadata>` + timestamp + `</metadata>`)},
		{"empty metadata with attributes", wrap(`<metadata xmlns:ext="urn:ext" />`), wrap(`<metadata xmlns:ext="urn:ext" >` + timestamp + `</metadata>`)},
		{"prefixed", `<c:bom xmlns:c="http://cyclonedx.org/schema/bom/1.4"><c:metadata/></c:bom>`, `<c:bom xmlns:c="http://cyclonedx.org/schema/bom/1.4"><c:metadata><c:timestamp>2026-10-22T10:34:56Z</c:timestamp></c:metadata></c:bom>`},
		{"remove reproducibility claim", wrap(`<metadata><properties><property name="keep">yes</property><property name="cdx:reproducible">enabled</property><property name="cdx:reproducible"/></properties></metadata>`), wrap(`<metadata>` + timestamp + `<properties><property name="keep">yes</property></properties></metadata>`)},
		{"only reproducibility claim", wrap(`<metadata><properties><property name="cdx:reproducible">enabled</property></properties></metadata>`), wrap(`<metadata>` + timestamp + `<properties></properties></metadata>`)},
		{"preserve extensions and graph", wrap("\n<!-- comment --><metadata><ext:data xmlns:ext=\"urn:ext\"><![CDATA[<raw>]]></ext:data></metadata>\n<components><component bom-ref=\"dep\"/></components><dependencies><dependency ref=\"root\"><dependency ref=\"dep\"/></dependency></dependencies>"), wrap("\n<!-- comment --><metadata>" + timestamp + "<ext:data xmlns:ext=\"urn:ext\"><![CDATA[<raw>]]></ext:data></metadata>\n<components><component bom-ref=\"dep\"/></components><dependencies><dependency ref=\"root\"><dependency ref=\"dep\"/></dependency></dependencies>")},
		{"ignore foreign timestamp and property", wrap(`<metadata><x:timestamp xmlns:x="urn:ext">other</x:timestamp><properties><x:property xmlns:x="urn:ext" name="cdx:reproducible">keep</x:property></properties></metadata>`), wrap(`<metadata>` + timestamp + `<x:timestamp xmlns:x="urn:ext">other</x:timestamp><properties><x:property xmlns:x="urn:ext" name="cdx:reproducible">keep</x:property></properties></metadata>`)},
		{"ignore component claim", wrap(`<metadata><component><properties><property name="cdx:reproducible">keep</property></properties></component></metadata>`), wrap(`<metadata>` + timestamp + `<component><properties><property name="cdx:reproducible">keep</property></properties></component></metadata>`)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			input := []byte(tc.input)
			original := bytes.Clone(input)
			output, changed, err := EnsureBOMTimestamp(input, now)
			if err != nil || !changed || string(output) != tc.want {
				t.Fatalf("got %s, changed=%v, err=%v; want %s", output, changed, err, tc.want)
			}
			if !bytes.Equal(input, original) {
				t.Fatal("mutated input buffer")
			}
			again, changed, err := EnsureBOMTimestamp(output, now.Add(time.Hour))
			if err != nil || changed || !bytes.Equal(again, output) {
				t.Fatalf("not idempotent: changed=%v, err=%v", changed, err)
			}
		})
	}
}

func TestEnsureBOMTimestampPreservesExisting(t *testing.T) {
	for _, value := range []string{"2026-01-01T00:00:00Z", "2026-01-01T12:34:56.123456789+02:00"} {
		input := []byte(`<?xml version="1.0"?><bom xmlns="http://cyclonedx.org/schema/bom/1.4"><metadata>` + "\n  <timestamp>" + value + "</timestamp>\n" + `<properties><property name="cdx:reproducible">enabled</property></properties></metadata><Signature xmlns="http://www.w3.org/2000/09/xmldsig#"/></bom>`)
		output, changed, err := EnsureBOMTimestamp(input, time.Now())
		if err != nil || changed || !bytes.Equal(input, output) {
			t.Fatalf("existing timestamp not preserved: changed=%v, err=%v", changed, err)
		}
	}
}

func TestEnsureBOMTimestampRejectsInvalidInput(t *testing.T) {
	wrap := func(metadata string) string {
		return `<bom xmlns="http://cyclonedx.org/schema/bom/1.4">` + metadata + `</bom>`
	}
	for _, tc := range []struct{ name, input string }{
		{"empty", ""},
		{"malformed", wrap(`<metadata>`)},
		{"missing metadata", wrap(``)},
		{"nested metadata", wrap(`<component><metadata/></component>`)},
		{"multiple metadata", wrap(`<metadata/><metadata/>`)},
		{"multiple roots", wrap(`<metadata/>`) + wrap(`<metadata/>`)},
		{"wrong namespace", `<bom xmlns="http://cyclonedx.org/schema/bom/1.6"><metadata/></bom>`},
		{"no namespace", `<bom><metadata/></bom>`},
		{"wrong root", `<not-bom xmlns="http://cyclonedx.org/schema/bom/1.4"><metadata/></not-bom>`},
		{"foreign metadata", wrap(`<metadata xmlns="urn:ext"/>`)},
		{"duplicate timestamps", wrap(`<metadata><timestamp>2026-01-01T00:00:00Z</timestamp><timestamp>2026-01-01T00:00:00Z</timestamp></metadata>`)},
		{"empty timestamp", wrap(`<metadata><timestamp/></metadata>`)},
		{"invalid timestamp", wrap(`<metadata><timestamp>yesterday</timestamp></metadata>`)},
		{"timestamp with child", wrap(`<metadata><timestamp>2026-01-01T00:00:00Z<x/></timestamp></metadata>`)},
		{"text outside root", "unexpected" + wrap(`<metadata/>`)},
		{"DTD", `<!DOCTYPE bom [<!ENTITY x SYSTEM "file:///etc/passwd">]>` + wrap(`<metadata/>`)},
		{"signed", wrap(`<metadata/><Signature xmlns="http://www.w3.org/2000/09/xmldsig#"/>`)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			output, changed, err := EnsureBOMTimestamp([]byte(tc.input), time.Now())
			if err == nil || changed || output != nil {
				t.Fatalf("expected failure without output: changed=%v, err=%v, output=%s", changed, err, output)
			}
		})
	}
}

func FuzzEnsureBOMTimestamp(f *testing.F) {
	f.Add([]byte(`<bom xmlns="http://cyclonedx.org/schema/bom/1.4"><metadata/></bom>`))
	f.Add([]byte(`<bom xmlns="http://cyclonedx.org/schema/bom/1.4"><metadata><properties><property name="cdx:reproducible"/></properties></metadata></bom>`))
	f.Fuzz(func(t *testing.T, input []byte) {
		now := time.Date(2026, 10, 22, 0, 0, 0, 0, time.UTC)
		output, changed, err := EnsureBOMTimestamp(input, now)
		if err != nil {
			if changed || output != nil {
				t.Fatal("failure returned changed output")
			}
			return
		}
		if !changed && !bytes.Equal(input, output) {
			t.Fatal("unchanged output differs")
		}
		if changed && !strings.Contains(string(output), "2026-10-22T00:00:00Z") {
			t.Fatal("missing inserted timestamp")
		}
		again, changed, err := EnsureBOMTimestamp(output, now)
		if err != nil || changed || !bytes.Equal(output, again) {
			t.Fatalf("not idempotent: %v", err)
		}
	})
}
