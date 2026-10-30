//go:build unit
// +build unit

package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/mock"

	"github.com/stretchr/testify/assert"
)

var cpe mavenBuildCommonPipelineEnvironment

func TestMavenBuild(t *testing.T) {
	SetConfigOptions(ConfigCommandOptions{
		OpenFile: config.OpenPiperFile,
	})

	t.Run("mavenBuild should install the artifact", func(t *testing.T) {
		mockedUtils := newMavenMockUtils()

		options := mavenBuildOptions{}

		err := runMavenBuild(&options, nil, &mockedUtils, &cpe)

		assert.Nil(t, err)
		if assert.Equal(t, 1, len(mockedUtils.Calls), "Expected one maven invocation for the main build") {
			params := mockedUtils.Calls[0].Params
			assert.Equal(t, "mvn", mockedUtils.Calls[0].Exec)
			assert.Contains(t, params, "install", "Call should contain install goal")
			assert.Contains(t, params, "--activate-profiles")
			profileIdx := -1
			for i, p := range params {
				if p == "--activate-profiles" {
					profileIdx = i
					break
				}
			}
			if assert.Greater(t, profileIdx, -1, "--activate-profiles flag must be present") {
				assert.Equal(t, "!snapshot.build,!milestone.build,release.build", params[profileIdx+1])
			}
		}
	})

	t.Run("mavenBuild should verify (not install) when Verify is true", func(t *testing.T) {
		mockedUtils := newMavenMockUtils()

		options := mavenBuildOptions{Verify: true}

		err := runMavenBuild(&options, nil, &mockedUtils, &cpe)

		assert.Nil(t, err)
		if assert.Equal(t, 1, len(mockedUtils.Calls), "Expected one maven invocation for the verify build") {
			params := mockedUtils.Calls[0].Params
			assert.Equal(t, "mvn", mockedUtils.Calls[0].Exec)
			assert.Contains(t, params, "verify", "Call should contain verify goal")
			assert.NotContains(t, params, "install", "Call must not contain install goal when Verify is true")
			assert.Contains(t, params, "--activate-profiles")
			profileIdx := -1
			for i, p := range params {
				if p == "--activate-profiles" {
					profileIdx = i
					break
				}
			}
			if assert.Greater(t, profileIdx, -1, "--activate-profiles flag must be present") {
				assert.Equal(t, "!snapshot.build,!milestone.build,release.build", params[profileIdx+1])
			}
		}
	})

	t.Run("mavenBuild accepts profiles", func(t *testing.T) {
		mockedUtils := newMavenMockUtils()

		options := mavenBuildOptions{Profiles: []string{"profile1", "profile2"}}

		err := runMavenBuild(&options, nil, &mockedUtils, &cpe)

		assert.Nil(t, err)
		if assert.Equal(t, 1, len(mockedUtils.Calls), "Expected one maven invocation for the main build") {
			assert.Contains(t, mockedUtils.Calls[0].Params, "--activate-profiles")
			assert.Contains(t, mockedUtils.Calls[0].Params, "!snapshot.build,!milestone.build,release.build,profile1,profile2")
		}
	})

	t.Run("mavenBuild milestone opt-in suppresses !milestone.build and release.build builtins", func(t *testing.T) {
		// When the caller opts in to milestone.build the binary must NOT prepend
		// !milestone.build (Maven 3.x deactivation wins the clash) or release.build
		// (which would activate the wrong quality repo).  Only !snapshot.build is kept.
		// This mirrors the profile list that executeBuild.groovy / sapPiperStageCentralBuild.groovy
		// produce for Milestone quality: ['milestone.build'] with no release-quality prefix.
		mockedUtils := newMavenMockUtils()

		options := mavenBuildOptions{Profiles: []string{"milestone.build"}}

		err := runMavenBuild(&options, nil, &mockedUtils, &cpe)

		assert.Nil(t, err)
		if assert.Equal(t, 1, len(mockedUtils.Calls), "Expected one maven invocation for the main build") {
			assert.Contains(t, mockedUtils.Calls[0].Params, "--activate-profiles")
			assert.Contains(t, mockedUtils.Calls[0].Params, "!snapshot.build,milestone.build")
			assert.NotContains(t, mockedUtils.Calls[0].Params, "!milestone.build",
				"!milestone.build must not be prepended when the caller opts in to milestone.build")
			assert.NotContains(t, mockedUtils.Calls[0].Params, "release.build",
				"release.build must not be prepended for a milestone-quality build")
		}
	})

	t.Run("mavenBuild BOM with milestone opt-in suppresses !milestone.build and release.build", func(t *testing.T) {
		// Both the aggregate-BOM invocation and the separate makeBom invocation must use the
		// milestone-aware profile set when milestone.build is present in user profiles.
		mockedUtils := newMavenMockUtils()

		options := mavenBuildOptions{CreateBOM: true, Profiles: []string{"milestone.build"}}

		err := runMavenBuild(&options, nil, &mockedUtils, &cpe)

		assert.Nil(t, err)
		if assert.Equal(t, 2, len(mockedUtils.Calls), "Expected two Maven invocations") {
			assert.Contains(t, mockedUtils.Calls[0].Params, "!snapshot.build,milestone.build")
			assert.Contains(t, mockedUtils.Calls[1].Params, "!snapshot.build,milestone.build")
			assert.NotContains(t, mockedUtils.Calls[0].Params, "!milestone.build")
			assert.NotContains(t, mockedUtils.Calls[1].Params, "!milestone.build")
		}
	})

	t.Run("mavenBuild explicit opt-in to snapshot does not affect milestone exclusion logic", func(t *testing.T) {
		// When the caller opts in to snapshot.build the binary must NOT prepend
		// !snapshot.build (Maven 3.x deactivation wins the clash), so the assembled
		// profile string must be '!milestone.build,release.build,snapshot.build'.
		// The !snapshot.build deactivation must be absent so that snapshot resolution
		// is actually enabled.
		mockedUtils := newMavenMockUtils()

		options := mavenBuildOptions{Profiles: []string{"snapshot.build"}}

		err := runMavenBuild(&options, nil, &mockedUtils, &cpe)

		assert.Nil(t, err)
		if assert.Equal(t, 1, len(mockedUtils.Calls), "Expected one maven invocation for the main build") {
			assert.Contains(t, mockedUtils.Calls[0].Params, "--activate-profiles")
			assert.Contains(t, mockedUtils.Calls[0].Params, "!milestone.build,release.build,snapshot.build",
				"snapshot opt-in must yield !milestone.build,release.build,snapshot.build — NOT the default Release set")
			assert.NotContains(t, mockedUtils.Calls[0].Params, "!snapshot.build,!milestone.build,release.build,snapshot.build",
				"!snapshot.build must be absent when the caller opts in to snapshot.build (Maven 3.x deactivation would otherwise win)")
		}
	})

	t.Run("mavenBuild should create BOM", func(t *testing.T) {
		mockedUtils := newMavenMockUtils()

		options := mavenBuildOptions{CreateBOM: true}

		err := runMavenBuild(&options, nil, &mockedUtils, &cpe)

		assert.Nil(t, err)
		if assert.Equal(t, 2, len(mockedUtils.Calls), "Expected two Maven invocations (default + makeAggregateBom)") {
			assert.Equal(t, "mvn", mockedUtils.Calls[1].Exec)
			assert.Contains(t, mockedUtils.Calls[0].Params, mvnCycloneDXPackage+":makeAggregateBom")
			assert.Contains(t, mockedUtils.Calls[0].Params, "-DoutputName=bom-maven")
		}
	})

	t.Run("mavenBuild BOM invocations carry deactivation profiles", func(t *testing.T) {
		mockedUtils := newMavenMockUtils()

		options := mavenBuildOptions{CreateBOM: true}

		err := runMavenBuild(&options, nil, &mockedUtils, &cpe)

		assert.Nil(t, err)
		if assert.Equal(t, 2, len(mockedUtils.Calls), "Expected two Maven invocations") {
			// First call: makeAggregateBom bundled with install
			assert.Contains(t, mockedUtils.Calls[0].Params, "--activate-profiles")
			assert.Contains(t, mockedUtils.Calls[0].Params, "!snapshot.build,!milestone.build,release.build")
			// Second call: runMakeBOMGoal (makeBom)
			assert.Contains(t, mockedUtils.Calls[1].Params, "--activate-profiles")
			assert.Contains(t, mockedUtils.Calls[1].Params, "!snapshot.build,!milestone.build,release.build")
		}
	})

	t.Run("mavenBuild BOM invocations merge user profiles with deactivation flags", func(t *testing.T) {
		mockedUtils := newMavenMockUtils()

		options := mavenBuildOptions{CreateBOM: true, Profiles: []string{"my-profile"}}

		err := runMavenBuild(&options, nil, &mockedUtils, &cpe)

		assert.Nil(t, err)
		if assert.Equal(t, 2, len(mockedUtils.Calls), "Expected two Maven invocations") {
			assert.Contains(t, mockedUtils.Calls[0].Params, "!snapshot.build,!milestone.build,release.build,my-profile")
			assert.Contains(t, mockedUtils.Calls[1].Params, "!snapshot.build,!milestone.build,release.build,my-profile")
		}
	})

	t.Run("mavenBuild deploy invocation does not carry build-phase deactivation profiles", func(t *testing.T) {
		// By design the deploy invocation replaces Flags with deployFlags and does NOT
		// carry '--activate-profiles !snapshot.build,...'.  Dependency resolution was
		// completed by the preceding 'mvn install'; deploy only transfers already-built
		// artifacts.  This intentional omission is documented in the code comment at
		// the mavenOptions.Flags assignment and in the mavenBuild.yaml longDescription.
		mockedUtils := newMavenMockUtils()

		options := mavenBuildOptions{Publish: true, Verify: false, AltDeploymentRepositoryID: "ID", AltDeploymentRepositoryURL: "http://sampleRepo.com", AltDeploymentRepositoryUser: "user", AltDeploymentRepositoryPassword: "pass"}

		err := runMavenBuild(&options, nil, &mockedUtils, &cpe)

		assert.Nil(t, err)
		if assert.Equal(t, 2, len(mockedUtils.Calls), "Expected two Maven invocations (main and deploy)") {
			// The deploy call overwrites Flags with deployFlags and must not carry the
			// build-phase deactivation profile argument. Check the actual combined-string
			// element rather than individual tokens (which are never standalone slice
			// elements and would give a vacuous true).
			assert.NotContains(t, mockedUtils.Calls[1].Params, "!snapshot.build,!milestone.build,release.build",
				"deploy call must not carry build-phase deactivation profiles")
			// Positive assertions: the deploy call must carry the deploy goal and the
			// configured alt-deployment-repository flag, and must not include
			// --activate-profiles (that flag only appears in the build-phase Flags).
			assert.Contains(t, mockedUtils.Calls[1].Params, "deploy",
				"deploy invocation must carry the deploy goal")
			assert.NotContains(t, mockedUtils.Calls[1].Params, "--activate-profiles",
				"deploy invocation must not carry --activate-profiles")
			assert.Contains(t, mockedUtils.Calls[1].Params, "-DaltDeploymentRepository=ID::http://sampleRepo.com",
				"deploy invocation must carry the configured alt-deployment-repository")
		}
	})

	t.Run("mavenBuild deploy with milestone.build uses milestone profile on build and no profile on deploy", func(t *testing.T) {
		// Verifies the combination of a milestone.build opt-in and Publish:true.
		// (a) The main build call must use the milestone-aware builtin profile set
		//     ("!snapshot.build,milestone.build") and must NOT include !milestone.build
		//     or release.build in the profile string.
		// (b) The deploy call must not carry any build-phase profile flags at all.
		mockedUtils := newMavenMockUtils()

		options := mavenBuildOptions{
			Profiles:                        []string{"milestone.build"},
			Publish:                         true,
			Verify:                          false,
			AltDeploymentRepositoryID:       "ID",
			AltDeploymentRepositoryURL:      "http://sampleRepo.com",
			AltDeploymentRepositoryUser:     "user",
			AltDeploymentRepositoryPassword: "pass",
		}

		err := runMavenBuild(&options, nil, &mockedUtils, &cpe)

		assert.Nil(t, err)
		if assert.Equal(t, 2, len(mockedUtils.Calls), "Expected two Maven invocations (main build and deploy)") {
			// (a) main build must use the milestone-aware profile set
			assert.Contains(t, mockedUtils.Calls[0].Params, "!snapshot.build,milestone.build",
				"main build must use the milestone-aware builtin profile set")
			assert.NotContains(t, mockedUtils.Calls[0].Params, "!milestone.build",
				"main build must not deactivate milestone.build when the caller opts in")
			assert.NotContains(t, mockedUtils.Calls[0].Params, "release.build",
				"release.build must not be prepended for a milestone-quality build")
			// (b) deploy call must not carry any build-phase profile flags
			assert.Contains(t, mockedUtils.Calls[1].Params, "deploy",
				"second invocation must carry the deploy goal")
			assert.NotContains(t, mockedUtils.Calls[1].Params, "--activate-profiles",
				"deploy call must not carry --activate-profiles")
			assert.NotContains(t, mockedUtils.Calls[1].Params, "!snapshot.build,milestone.build",
				"deploy call must not forward the build-phase profile string")
		}
	})

	t.Run("mavenBuild include install and deploy when publish is true", func(t *testing.T) {
		mockedUtils := newMavenMockUtils()

		options := mavenBuildOptions{Publish: true, Verify: false, AltDeploymentRepositoryID: "ID", AltDeploymentRepositoryURL: "http://sampleRepo.com", AltDeploymentRepositoryUser: "user", AltDeploymentRepositoryPassword: "pass"}

		err := runMavenBuild(&options, nil, &mockedUtils, &cpe)

		assert.Nil(t, err)
		if assert.Equal(t, 2, len(mockedUtils.Calls), "Expected two Maven invocations (main and deploy)") {
			assert.Contains(t, mockedUtils.Calls[0].Params, "install")
			assert.NotContains(t, mockedUtils.Calls[0].Params, "verify")
			assert.Contains(t, mockedUtils.Calls[1].Params, "deploy")
		}
	})

	t.Run("mavenBuild with deploy must skip build, install and test", func(t *testing.T) {
		mockedUtils := newMavenMockUtils()

		options := mavenBuildOptions{Publish: true, Verify: false, DeployFlags: []string{"-Dmaven.main.skip=true", "-Dmaven.test.skip=true", "-Dmaven.install.skip=true"}, AltDeploymentRepositoryID: "ID", AltDeploymentRepositoryURL: "http://sampleRepo.com", AltDeploymentRepositoryUser: "user", AltDeploymentRepositoryPassword: "pass"}

		err := runMavenBuild(&options, nil, &mockedUtils, &cpe)

		assert.Nil(t, err)
		if assert.Equal(t, 2, len(mockedUtils.Calls), "Expected two Maven invocations (main and deploy)") {
			assert.Contains(t, mockedUtils.Calls[1].Params, "-Dmaven.main.skip=true")
			assert.Contains(t, mockedUtils.Calls[1].Params, "-Dmaven.test.skip=true")
			assert.Contains(t, mockedUtils.Calls[1].Params, "-Dmaven.install.skip=true")
		}
	})

	t.Run("mavenBuild with deploy must include alt repo id and url when passed as parameter", func(t *testing.T) {
		mockedUtils := newMavenMockUtils()

		options := mavenBuildOptions{Publish: true, Verify: false, AltDeploymentRepositoryID: "ID", AltDeploymentRepositoryURL: "http://sampleRepo.com", AltDeploymentRepositoryUser: "user", AltDeploymentRepositoryPassword: "pass"}

		err := runMavenBuild(&options, nil, &mockedUtils, &cpe)

		assert.Nil(t, err)
		if assert.Equal(t, 2, len(mockedUtils.Calls), "Expected two Maven invocations (main and deploy)") {
			assert.Contains(t, mockedUtils.Calls[1].Params, "-DaltDeploymentRepository=ID::http://sampleRepo.com")
		}
	})

	t.Run("mavenBuild with deploy must not set altDeploymentRepository when URL is missing", func(t *testing.T) {
		mockedUtils := newMavenMockUtils()

		options := mavenBuildOptions{Publish: true, Verify: false, AltDeploymentRepositoryID: "ID"}

		err := runMavenBuild(&options, nil, &mockedUtils, &cpe)

		assert.Nil(t, err)
		if assert.Equal(t, 2, len(mockedUtils.Calls), "Expected two Maven invocations (main and deploy)") {
			assert.NotContains(t, mockedUtils.Calls[1].Params, "-DaltDeploymentRepository=ID::")
		}
	})

	t.Run("mavenBuild with deploy must not set altDeploymentRepository when ID is missing", func(t *testing.T) {
		mockedUtils := newMavenMockUtils()

		options := mavenBuildOptions{Publish: true, Verify: false, AltDeploymentRepositoryURL: "http://sampleRepo.com"}

		err := runMavenBuild(&options, nil, &mockedUtils, &cpe)

		assert.Nil(t, err)
		if assert.Equal(t, 2, len(mockedUtils.Calls), "Expected two Maven invocations (main and deploy)") {
			assert.NotContains(t, mockedUtils.Calls[1].Params, "-DaltDeploymentRepository=::http://sampleRepo.com")
		}
	})

	t.Run("mavenBuild must not set altDeploymentRepository when Publish is false", func(t *testing.T) {
		mockedUtils := newMavenMockUtils()

		options := mavenBuildOptions{Publish: false, AltDeploymentRepositoryID: "ID", AltDeploymentRepositoryURL: "http://sampleRepo.com"}

		err := runMavenBuild(&options, nil, &mockedUtils, &cpe)

		assert.Nil(t, err)
		if assert.Equal(t, 1, len(mockedUtils.Calls), "Expected one Maven invocation (no deploy when Publish is false)") {
			assert.NotContains(t, mockedUtils.Calls[0].Params, "-DaltDeploymentRepository=ID::http://sampleRepo.com")
		}
	})

	t.Run("mavenBuild should not create build artifacts metadata when CreateBuildArtifactsMetadata is false and Publish is true", func(t *testing.T) {
		mockedUtils := newMavenMockUtils()
		mockedUtils.AddFile("pom.xml", []byte{})
		options := mavenBuildOptions{CreateBuildArtifactsMetadata: false, Publish: true, AltDeploymentRepositoryID: "ID", AltDeploymentRepositoryURL: "http://sampleRepo.com", AltDeploymentRepositoryUser: "user", AltDeploymentRepositoryPassword: "pass"}
		err := runMavenBuild(&options, nil, &mockedUtils, &cpe)
		assert.Nil(t, err)
		assert.Equal(t, mockedUtils.Calls[0].Exec, "mvn")
		assert.Contains(t, mockedUtils.Calls[0].Params, "install")
		assert.Empty(t, cpe.custom.mavenBuildArtifacts)
	})

	t.Run("mavenBuild should not create build artifacts metadata when CreateBuildArtifactsMetadata is true and Publish is false", func(t *testing.T) {
		mockedUtils := newMavenMockUtils()
		mockedUtils.AddFile("pom.xml", []byte{})
		options := mavenBuildOptions{CreateBuildArtifactsMetadata: true, Publish: false}
		err := runMavenBuild(&options, nil, &mockedUtils, &cpe)
		assert.Nil(t, err)
		assert.Equal(t, mockedUtils.Calls[0].Exec, "mvn")
		assert.Empty(t, cpe.custom.mavenBuildArtifacts)
	})
}

// TestContainsProfile exercises the containsProfile helper directly so that its
// exact-match contract is pinned and regressions are caught without going through
// the full runMavenBuild invocation path.
func TestContainsProfile(t *testing.T) {
	tests := []struct {
		name     string
		profiles []string
		target   string
		want     bool
	}{
		{
			name:     "empty slice returns false",
			profiles: []string{},
			target:   "milestone.build",
			want:     false,
		},
		{
			name:     "nil slice returns false",
			profiles: nil,
			target:   "milestone.build",
			want:     false,
		},
		{
			name:     "single-element match returns true",
			profiles: []string{"milestone.build"},
			target:   "milestone.build",
			want:     true,
		},
		{
			name:     "match in multi-element slice returns true",
			profiles: []string{"foo", "milestone.build", "bar"},
			target:   "milestone.build",
			want:     true,
		},
		{
			name:     "no match in multi-element slice returns false",
			profiles: []string{"foo", "bar"},
			target:   "milestone.build",
			want:     false,
		},
		{
			name:     "partial string does not match",
			profiles: []string{"milestone.build.extra"},
			target:   "milestone.build",
			want:     false,
		},
		{
			name:     "prefix does not match",
			profiles: []string{"milestone"},
			target:   "milestone.build",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := containsProfile(tt.profiles, tt.target)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestBuildBuiltinProfiles exercises the buildBuiltinProfiles helper directly.
// The table covers every branch in the function so that regressions are caught
// before they surface in the higher-level runMavenBuild tests.
func TestBuildBuiltinProfiles(t *testing.T) {
	tests := []struct {
		name         string
		userProfiles []string
		want         []string
	}{
		{
			// Default path: no special profiles → full guard set.
			name:         "empty input returns default release profile set",
			userProfiles: []string{},
			want:         []string{"!snapshot.build", "!milestone.build", "release.build"},
		},
		{
			name:         "nil input returns default release profile set",
			userProfiles: nil,
			want:         []string{"!snapshot.build", "!milestone.build", "release.build"},
		},
		{
			// milestone.build opt-in: only the snapshot exclusion is prepended.
			// "!milestone.build" and "release.build" are omitted so the caller's
			// milestone.build profile is not blocked by a conflicting deactivation.
			name:         "milestone.build opts in: only snapshot exclusion is prepended",
			userProfiles: []string{"milestone.build"},
			want:         []string{"!snapshot.build"},
		},
		{
			name:         "milestone.build alongside other profiles still gives snapshot-only exclusion",
			userProfiles: []string{"my-profile", "milestone.build"},
			want:         []string{"!snapshot.build"},
		},
		{
			// snapshot.build opt-in: only the milestone/release guards are prepended.
			// "!snapshot.build" is omitted so the caller's snapshot.build profile
			// actually enables snapshot dependency resolution (Maven 3.x deactivation
			// would win the !snapshot.build / snapshot.build clash, blocking access).
			name:         "snapshot.build opts in: snapshot exclusion is omitted",
			userProfiles: []string{"snapshot.build"},
			want:         []string{"!milestone.build", "release.build"},
		},
		{
			name:         "snapshot.build alongside other profiles still omits snapshot exclusion",
			userProfiles: []string{"my-profile", "snapshot.build"},
			want:         []string{"!milestone.build", "release.build"},
		},
		{
			// Both opt-ins present: milestone.build takes precedence (checked first).
			name:         "milestone.build and snapshot.build together: milestone.build wins",
			userProfiles: []string{"milestone.build", "snapshot.build"},
			want:         []string{"!snapshot.build"},
		},
		{
			// Unrelated multi-profile input → default set.
			name:         "multiple unrelated profiles return default set",
			userProfiles: []string{"profile1", "profile2"},
			want:         []string{"!snapshot.build", "!milestone.build", "release.build"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildBuiltinProfiles(tt.userProfiles)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestLoadRemoteRepoCertificates(t *testing.T) {
	t.Run("should find cacerts at Java 9+ path", func(t *testing.T) {
		filesMock := &mock.FilesMock{}
		execMock := &mock.ExecMockRunner{}

		javaHome := "/usr/lib/jvm/java-17"
		os.Setenv("JAVA_HOME", javaHome)
		defer os.Unsetenv("JAVA_HOME")

		java9Path := filepath.Join(javaHome, "lib", "security", "cacerts")
		filesMock.AddFile(java9Path, []byte("cacerts content"))
		filesMock.AddFile(".pipeline/mavenCaCerts", []byte{})

		var flags []string
		err := loadRemoteRepoCertificates([]string{}, nil, &flags, execMock, filesMock, "")

		assert.NoError(t, err)
	})

	t.Run("should fall back to Java 8 path", func(t *testing.T) {
		filesMock := &mock.FilesMock{}
		execMock := &mock.ExecMockRunner{}

		javaHome := "/usr/lib/jvm/java-8"
		os.Setenv("JAVA_HOME", javaHome)
		defer os.Unsetenv("JAVA_HOME")

		java8Path := filepath.Join(javaHome, "jre", "lib", "security", "cacerts")
		filesMock.AddFile(java8Path, []byte("cacerts content"))
		filesMock.AddFile(".pipeline/mavenCaCerts", []byte{})

		var flags []string
		err := loadRemoteRepoCertificates([]string{}, nil, &flags, execMock, filesMock, "")

		assert.NoError(t, err)
	})

	t.Run("should use custom javaCaCertFilePath when provided", func(t *testing.T) {
		filesMock := &mock.FilesMock{}
		execMock := &mock.ExecMockRunner{}

		customPath := "/custom/path/to/cacerts"
		filesMock.AddFile(customPath, []byte("cacerts content"))
		filesMock.AddFile(".pipeline/mavenCaCerts", []byte{})

		var flags []string
		err := loadRemoteRepoCertificates([]string{}, nil, &flags, execMock, filesMock, customPath)

		assert.NoError(t, err)
	})

	t.Run("should return nil and warn when cacerts not found", func(t *testing.T) {
		filesMock := &mock.FilesMock{}
		execMock := &mock.ExecMockRunner{}

		javaHome := "/usr/lib/jvm/java-17"
		os.Setenv("JAVA_HOME", javaHome)
		defer os.Unsetenv("JAVA_HOME")

		// Don't add any cacerts file - simulating missing file

		var flags []string
		err := loadRemoteRepoCertificates([]string{}, nil, &flags, execMock, filesMock, "")

		// Should not error - just warn and continue
		assert.NoError(t, err)
	})
}
