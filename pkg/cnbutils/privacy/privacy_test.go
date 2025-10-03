package privacy_test

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/cnbutils/privacy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCnbPrivacy_FilterBuildpacks(t *testing.T) {
	t.Parallel()

	t.Run("allows paketo", func(t *testing.T) {
		aliases := []string{
			"paketobuildpacks/nodejs:v1",
			"docker.io/paketobuildpacks/nodejs:v1",
			"index.docker.io/paketobuildpacks/nodejs:v1",
			"docker.io/paketobuildpacks/nodejs:v1",
		}

		filtered := privacy.FilterBuildpacks(aliases)

		require.Len(t, filtered, len(aliases))
		for i := range filtered {
			assert.Equal(t, aliases[i], filtered[i])
		}
	})

	t.Run("allows heroku", func(t *testing.T) {
		aliases := []string{
			"public.ecr.aws/heroku-buildpacks/heroku-jvm-buildpack@sha256:3a8ee9ebf88e47c5e30bc5712fb2794380aed75552499f92bd6773ec446421ef",
		}

		filtered := privacy.FilterBuildpacks(aliases)

		require.Len(t, filtered, len(aliases))
		for i := range filtered {
			assert.Equal(t, aliases[i], filtered[i])
		}
	})

	t.Run("allows google buildpacks", func(t *testing.T) {
		aliases := []string{
			"gcr.io/buildpacks/java:latest",
			"gcr.io/buildpacks/java",
		}

		filtered := privacy.FilterBuildpacks(aliases)

		require.Len(t, filtered, len(aliases))
		for i := range filtered {
			assert.Equal(t, aliases[i], filtered[i])
		}
	})

	t.Run("filters others", func(t *testing.T) {
		images := []string{
			"test/nodejs:v1",
			"test/nodejs:v1", // SHA should be the same for multiple occurences
			"my-mirror.de/paketobuildpacks/nodejs:v1",
			"gcr.io/my-project/paketo-buildpacks/nodejs:v1",
		}

		filtered := privacy.FilterBuildpacks(images)

		require.Len(t, filtered, len(images))

		assert.ElementsMatch(t, filtered, []string{
			"6ea013d746199ccc0e48e0b4984a6d9357105b82f936ecf18d15786805ac892f",
			"6ea013d746199ccc0e48e0b4984a6d9357105b82f936ecf18d15786805ac892f",
			"66131ef922cf26b1500e54a74827f051b43857bcf8d0596593c182548f7d4bd6",
			"4fd8f0a950aacd7e428c79fce6f51bb1fbf0ab15caf4aca7accc18609acd79b1",
		})
	})

	t.Run("fails gracefully on parse error", func(t *testing.T) {
		images := []string{
			"test/nodejs v1 spaces are not allowed",
		}

		filtered := privacy.FilterBuildpacks(images)

		require.Len(t, filtered, len(images))
		for _, image := range filtered {
			assert.Equal(t, "<error>", image)
		}
	})

}

func TestCnbPrivacy_FilterEnv(t *testing.T) {
	t.Parallel()

	t.Run("copies only allow listed keys", func(t *testing.T) {
		env := map[string]interface{}{
			"PRIVATE":         "paketobuildpacks/nodejs:v1",
			"BP_NODE_VERSION": "8",
			"BP_JVM_VERSION":  "11",
		}

		filteredEnv := privacy.FilterEnv(env)

		assert.Equal(t, map[string]interface{}{
			"BP_NODE_VERSION": "8",
			"BP_JVM_VERSION":  "11",
		}, filteredEnv)
	})

	t.Run("works on nil map", func(t *testing.T) {
		var env map[string]interface{} = nil

		filteredEnv := privacy.FilterEnv(env)

		assert.Empty(t, filteredEnv)
	})
}

func TestCnbPrivacy_FilterBuilder(t *testing.T) {
	t.Parallel()

	t.Run("allows paketo", func(t *testing.T) {
		builder := []string{
			"paketobuildpacks/builder:tiny",
			"paketobuildpacks/builder:base",
			"paketobuildpacks/builder:full",
		}

		for _, b := range builder {
			filteredBuilder := privacy.FilterBuilder(b)
			assert.Equal(t, b, filteredBuilder)
		}

	})

	t.Run("filters unknown builders", func(t *testing.T) {
		builder := "notpaketobuildpacks/builder:base"

		filteredBuilder := privacy.FilterBuilder(builder)

		assert.Equal(t, "70278d9360533fa4978e5c50aa79bc35a8c0167a353e00521202feeaa09a305b", filteredBuilder)
	})

}
