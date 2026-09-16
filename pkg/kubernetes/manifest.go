package kubernetes

import (
	"bytes"
	"io"
	"strings"

	"go.yaml.in/yaml/v3"
)

// ExtractImagesFromManifests parses a multi-doc YAML byte stream (as produced
// by `helm template`) and returns every non-placeholder value of any `image`
// key, deduplicated in first-seen order. It covers container/initContainer
// specs at any depth (Deployment, StatefulSet, DaemonSet, Job, CronJob, ...)
// by recursively walking each document rather than relying on typed structs.
//
// Decoding uses yaml.Node so that document/key order is preserved — a plain
// map[string]interface{} would lose ordering and make the result depend on
// Go's randomized map iteration.
func ExtractImagesFromManifests(manifests []byte) ([]string, error) {
	dec := yaml.NewDecoder(bytes.NewReader(manifests))
	seen := map[string]struct{}{}
	var out []string
	for {
		var doc yaml.Node
		if err := dec.Decode(&doc); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		walkNode(&doc, func(value string) {
			if value == "" || strings.Contains(value, "{{") {
				return
			}
			if _, ok := seen[value]; ok {
				return
			}
			seen[value] = struct{}{}
			out = append(out, value)
		})
	}
	return out, nil
}

// walkNode recursively traverses a yaml.Node tree in document order, invoking
// visitImage for every scalar value held under a mapping key named "image".
func walkNode(node *yaml.Node, visitImage func(value string)) {
	switch node.Kind {
	case yaml.DocumentNode:
		for _, child := range node.Content {
			walkNode(child, visitImage)
		}
	case yaml.MappingNode:
		// Mapping content is a flat [key0, val0, key1, val1, ...] slice.
		for i := 0; i+1 < len(node.Content); i += 2 {
			key := node.Content[i]
			val := node.Content[i+1]
			if key.Value == "image" && val.Kind == yaml.ScalarNode {
				visitImage(val.Value)
			}
			walkNode(val, visitImage)
		}
	case yaml.SequenceNode:
		for _, child := range node.Content {
			walkNode(child, visitImage)
		}
	}
}
