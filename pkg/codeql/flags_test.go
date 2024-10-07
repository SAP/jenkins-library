package codeql

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsFlagSetByUser(t *testing.T) {
	t.Parallel()

	customFlags := map[string]string{
		"--flag1": "--flag1=1",
		"-f2":     "-f2=2",
		"--flag3": "--flag3",
	}

	t.Run("Flag is not set by user", func(t *testing.T) {
		input := []string{"-f4"}
		assert.False(t, IsFlagSetByUser(customFlags, input))
	})
	t.Run("Flag is set by user", func(t *testing.T) {
		input := []string{"-f2"}
		assert.True(t, IsFlagSetByUser(customFlags, input))
	})
	t.Run("One of flags is set by user", func(t *testing.T) {
		input := []string{"--flag2", "-f2"}
		assert.True(t, IsFlagSetByUser(customFlags, input))
	})
}

func TestAppendFlagIfNotSetByUser(t *testing.T) {
	t.Parallel()

	t.Run("Flag is not set by user", func(t *testing.T) {
		result := []string{}
		flagsToCheck := []string{"--flag1", "-f1"}
		flagToAppend := []string{"--flag1=1"}
		customFlags := map[string]string{
			"--flag2": "--flag2=1",
		}
		result = AppendFlagIfNotSetByUser(result, flagsToCheck, flagToAppend, customFlags)
		assert.Equal(t, 1, len(result))
		assert.Equal(t, "--flag1=1", result[0])
	})

	t.Run("Flag is set by user", func(t *testing.T) {
		result := []string{}
		flagsToCheck := []string{"--flag1", "-f1"}
		flagToAppend := []string{"--flag1=1"}
		customFlags := map[string]string{
			"--flag1": "--flag1=2",
		}
		result = AppendFlagIfNotSetByUser(result, flagsToCheck, flagToAppend, customFlags)
		assert.Equal(t, 0, len(result))
	})
}

func TestAppendCustomFlags(t *testing.T) {
	t.Parallel()

	t.Run("Flags with values", func(t *testing.T) {
		flags := map[string]string{
			"--flag1": "--flag1=1",
			"--flag2": "--flag2=2",
			"--flag3": "--flag3=3",
		}
		result := []string{}
		result = AppendCustomFlags(result, flags)
		assert.Equal(t, 3, len(result))
		jointFlags := strings.Join(result, " ")
		assert.True(t, strings.Contains(jointFlags, "--flag1=1"))
		assert.True(t, strings.Contains(jointFlags, "--flag2=2"))
		assert.True(t, strings.Contains(jointFlags, "--flag3=3"))
	})
	t.Run("Flags without values", func(t *testing.T) {
		flags := map[string]string{
			"--flag1": "--flag1",
			"--flag2": "--flag2",
			"--flag3": "--flag3",
		}
		result := []string{}
		result = AppendCustomFlags(result, flags)
		assert.Equal(t, 3, len(result))
		jointFlags := strings.Join(result, " ")
		assert.True(t, strings.Contains(jointFlags, "--flag1"))
		assert.True(t, strings.Contains(jointFlags, "--flag2"))
		assert.True(t, strings.Contains(jointFlags, "--flag3"))
	})
	t.Run("Some flags without values", func(t *testing.T) {
		flags := map[string]string{
			"--flag1": "--flag1=1",
			"--flag2": "--flag2=1",
			"--flag3": "--flag3",
		}
		result := []string{}
		result = AppendCustomFlags(result, flags)
		assert.Equal(t, 3, len(result))
		jointFlags := strings.Join(result, " ")
		assert.True(t, strings.Contains(jointFlags, "--flag1=1"))
		assert.True(t, strings.Contains(jointFlags, "--flag2=1"))
		assert.True(t, strings.Contains(jointFlags, "--flag3"))
	})
	t.Run("Empty input", func(t *testing.T) {
		flags := map[string]string{}
		expected := []string{}
		result := []string{}
		result = AppendCustomFlags(result, flags)
		assert.Equal(t, expected, result)
	})
}

func TestParseFlags(t *testing.T) {
	t.Parallel()

	t.Run("Valid flags with values", func(t *testing.T) {
		inputStr := "--flag1=1 --flag2=2 --flag3=string"
		expected := map[string]bool{
			"--flag1=1":      true,
			"--flag2=2":      true,
			"--flag3=string": true,
		}
		result := parseFlags(inputStr)
		assert.Equal(t, len(expected), len(result))
		for _, f := range result {
			assert.True(t, expected[f])
		}
	})

	t.Run("Valid flags without values", func(t *testing.T) {
		inputStr := "--flag1 -flag2 -f3"
		expected := map[string]bool{
			"--flag1": true,
			"-flag2":  true,
			"-f3":     true,
		}
		result := parseFlags(inputStr)
		assert.Equal(t, len(expected), len(result))
		for _, f := range result {
			assert.True(t, expected[f])
		}
	})

	t.Run("Valid flags with spaces in value", func(t *testing.T) {
		inputStr := "--flag1='mvn install' --flag2=\"mvn clean install\" -f3='mvn clean install -DskipTests=true'"
		expected := map[string]bool{
			"--flag1=mvn install":                    true,
			"--flag2=mvn clean install":              true,
			"-f3=mvn clean install -DskipTests=true": true,
		}
		result := parseFlags(inputStr)
		assert.Equal(t, len(expected), len(result))
		for _, f := range result {
			assert.True(t, expected[f])
		}
	})
}

func TestRemoveDuplicateFlags(t *testing.T) {
	t.Parallel()

	longShortFlags := map[string]string{
		"--flag1": "-f1",
		"--flag2": "-f2",
		"--flag3": "-f3",
	}

	t.Run("No duplications", func(t *testing.T) {
		flags := map[string]string{
			"--flag1": "--flag1=1",
			"-f2":     "-f2=2",
			"--flag3": "--flag3",
		}
		expected := map[string]string{
			"--flag1": "--flag1=1",
			"-f2":     "-f2=2",
			"--flag3": "--flag3",
		}
		removeDuplicateFlags(flags, longShortFlags)
		assert.Equal(t, len(expected), len(flags))
		for k, v := range flags {
			assert.Equal(t, expected[k], v)
		}
	})

	t.Run("Duplications", func(t *testing.T) {
		flags := map[string]string{
			"--flag1": "--flag1=1",
			"-f1":     "-f1=2",
			"--flag2": "--flag2=1",
			"-f2":     "-f2=2",
			"--flag3": "--flag3",
			"-f3":     "-f3",
		}
		expected := map[string]string{
			"--flag1": "--flag1=1",
			"--flag2": "--flag2=1",
			"--flag3": "--flag3",
		}
		removeDuplicateFlags(flags, longShortFlags)
		assert.Equal(t, len(expected), len(flags))
		for k, v := range flags {
			assert.Equal(t, expected[k], v)
		}
	})
}

func TestParseCustomFlags(t *testing.T) {
	t.Parallel()

	t.Run("Valid flags with values", func(t *testing.T) {
		inputStr := "--flag1=1 --flag2=2 --flag3=string"
		expected := map[string]bool{
			"--flag1=1":      true,
			"--flag2=2":      true,
			"--flag3=string": true,
		}
		result := ParseCustomFlags(inputStr)
		assert.Equal(t, len(expected), len(result))
		for _, f := range result {
			assert.True(t, expected[f])
		}
	})

	t.Run("Valid flags with duplication", func(t *testing.T) {
		inputStr := "--flag1=1 --flag2=2 --flag3=string --flag2=3"
		expected := map[string]bool{
			"--flag1=1":      true,
			"--flag2=3":      true,
			"--flag3=string": true,
		}
		result := ParseCustomFlags(inputStr)
		assert.Equal(t, len(expected), len(result))
		for _, f := range result {
			assert.True(t, expected[f])
		}
	})

	t.Run("Valid flags with duplicated short flag", func(t *testing.T) {
		inputStr := "--flag1=1 --flag2=2 --flag3=string --language=java -l=python"
		expected := map[string]bool{
			"--flag1=1":       true,
			"--flag2=2":       true,
			"--flag3=string":  true,
			"--language=java": true,
		}
		result := ParseCustomFlags(inputStr)
		assert.Equal(t, len(expected), len(result))
		for _, f := range result {
			assert.True(t, expected[f])
		}
	})

	t.Run("Valid flags without values", func(t *testing.T) {
		inputStr := "--flag1 -flag2 -f3"
		expected := map[string]bool{
			"--flag1": true,
			"-flag2":  true,
			"-f3":     true,
		}
		result := ParseCustomFlags(inputStr)
		assert.Equal(t, len(expected), len(result))
		for _, f := range result {
			assert.True(t, expected[f])
		}
	})

	t.Run("Valid flags with spaces in value", func(t *testing.T) {
		inputStr := "--flag1='mvn install' --flag2=\"mvn clean install\" -f3='mvn clean install -DskipTests=true'"
		expected := map[string]bool{
			"--flag1=mvn install":                    true,
			"--flag2=mvn clean install":              true,
			"-f3=mvn clean install -DskipTests=true": true,
		}
		result := ParseCustomFlags(inputStr)
		assert.Equal(t, len(expected), len(result))
		for _, f := range result {
			assert.True(t, expected[f])
		}
	})
}

func TestAppendThreadsAndRam(t *testing.T) {
	t.Parallel()

	threads := "0"
	ram := "2000"

	t.Run("Threads and ram are set by user", func(t *testing.T) {
		customFlags := map[string]string{
			"--threads": "--threads=1",
			"--ram":     "--ram=3000",
		}
		params := []string{}
		params = AppendThreadsAndRam(params, threads, ram, customFlags)
		assert.Equal(t, 0, len(params))
	})

	t.Run("Threads and ram are not set by user", func(t *testing.T) {
		customFlags := map[string]string{}
		params := []string{}
		params = AppendThreadsAndRam(params, threads, ram, customFlags)
		assert.Equal(t, 2, len(params))
		paramsStr := strings.Join(params, " ")
		assert.True(t, strings.Contains(paramsStr, "--threads=0"))
		assert.True(t, strings.Contains(paramsStr, "--ram=2000"))
	})

	t.Run("Threads is set by user, ram is not", func(t *testing.T) {
		customFlags := map[string]string{
			"--threads": "--threads=1",
		}
		params := []string{}
		params = AppendThreadsAndRam(params, threads, ram, customFlags)
		assert.Equal(t, 1, len(params))
		assert.True(t, strings.Contains(params[0], "--ram=2000"))
	})

	t.Run("Add params to non-empty slice", func(t *testing.T) {
		customFlags := map[string]string{}
		params := []string{"cmd"}
		params = AppendThreadsAndRam(params, threads, ram, customFlags)
		assert.Equal(t, 3, len(params))
		assert.Equal(t, "cmd", params[0])
		assert.Equal(t, "--threads=0", params[1])
		assert.Equal(t, "--ram=2000", params[2])
	})
}
