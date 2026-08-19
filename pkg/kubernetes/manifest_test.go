//go:build unit
// +build unit

package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v3"
)

func TestRunHelmExtractImagesFromManifests(t *testing.T) {
	tests := []struct {
		name      string
		manifests string
		expected  []string
		wantErr   bool
	}{
		{
			name: "deployment with containers and initContainers",
			manifests: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
        - name: app
          image: registry.example.com/app:1.2.3
      initContainers:
        - name: init
          image: busybox:1.36
`,
			expected: []string{"registry.example.com/app:1.2.3", "busybox:1.36"},
		},
		{
			name: "multiple documents are traversed in order and deduplicated",
			manifests: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: first
spec:
  template:
    spec:
      containers:
        - name: app
          image: registry.example.com/app:1.0.0
---
apiVersion: batch/v1
kind: CronJob
metadata:
  name: second
spec:
  jobTemplate:
    spec:
      template:
        spec:
          containers:
            - name: cron
              image: registry.example.com/cron:2.0.0
            - name: app-again
              image: registry.example.com/app:1.0.0
`,
			expected: []string{"registry.example.com/app:1.0.0", "registry.example.com/cron:2.0.0"},
		},
		{
			name: "placeholders, empty values and templated refs are skipped",
			manifests: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
        - name: empty
          image: ""
        - name: templated
          image: "{{ .Values.image }}"
        - name: real
          image: registry.example.com/real:1.0.0
`,
			expected: []string{"registry.example.com/real:1.0.0"},
		},
		{
			name:      "no image keys yields an empty result",
			manifests: "apiVersion: v1\nkind: ConfigMap\ndata:\n  key: value\n",
			expected:  nil,
		},
		{
			name:      "empty input yields an empty result",
			manifests: "",
			expected:  nil,
		},
		{
			name:      "malformed YAML returns an error",
			manifests: "foo: [unclosed",
			wantErr:   true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			images, err := ExtractImagesFromManifests([]byte(test.manifests))
			if test.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, test.expected, images)
		})
	}
}

func TestRunHelmWalkNode(t *testing.T) {
	// collect gathers every image value walkNode visits, in order.
	collect := func(yamlDoc string) []string {
		var node yaml.Node
		require.NoError(t, yaml.Unmarshal([]byte(yamlDoc), &node))
		var got []string
		walkNode(&node, func(value string) { got = append(got, value) })
		return got
	}

	t.Run("visits image scalars under a mapping in document order", func(t *testing.T) {
		got := collect(`
image: a:1
nested:
  image: b:2
list:
  - image: c:3
  - image: d:4
`)
		assert.Equal(t, []string{"a:1", "b:2", "c:3", "d:4"}, got)
	})

	t.Run("ignores non-image keys and non-scalar image values", func(t *testing.T) {
		got := collect(`
name: notanimage
image:
  nested: notvisited
`)
		// "image" whose value is a mapping (not a scalar) must not be reported.
		assert.Empty(t, got)
	})

	t.Run("a scalar document node is a safe no-op", func(t *testing.T) {
		got := collect(`justascalar`)
		assert.Empty(t, got, "unhandled node kinds (scalar/alias) must not panic and yield nothing")
	})

	t.Run("a nil node is a safe no-op", func(t *testing.T) {
		var got []string
		walkNode(&yaml.Node{}, func(value string) { got = append(got, value) })
		assert.Empty(t, got)
	})
}
