package codeql

import (
	"strings"

	"github.com/SAP/jenkins-library/pkg/log"
)

var longShortFlagsMap = map[string]string{
	"--language":          "-l",
	"--command":           "-c",
	"--source-root":       "-s",
	"--github-url":        "-g",
	"--mode":              "-m",
	"--extractor-option":  "-O",
	"--github-auth-stdin": "-a",
	"--threads":           "-j",
	"--ram":               "-M",
}

func IsFlagSetByUser(customFlags map[string]string, flagsToCheck []string) bool {
	for _, flag := range flagsToCheck {
		if _, exists := customFlags[flag]; exists {
			return true
		}
	}
	return false
}

func AppendFlagIfNotSetByUser(cmd []string, flagToCheck []string, flagToAppend []string, customFlags map[string]string) []string {
	if !IsFlagSetByUser(customFlags, flagToCheck) {
		cmd = append(cmd, flagToAppend...)
	}
	return cmd
}

func AppendCustomFlags(cmd []string, flags map[string]string) []string {
	for _, flag := range flags {
		if strings.TrimSpace(flag) != "" {
			cmd = append(cmd, flag)
		}
	}
	return cmd
}

// parseFlags parses the input string and extracts individual flags.
// Flags can have values, which can be enclosed in single quotes (”) or double quotes ("").
// Flags are separated by whitespace.
// The function returns a slice of strings, where each string represents a flag or a flag with its value.
func parseFlags(input string) []string {
	result := []string{}
	isFlagStarted := false
	isString := false
	flag := ""
	for i, c := range input {
		if !isFlagStarted {
			if string(c) == " " {
				continue
			}
			flag += string(c)
			isFlagStarted = true
			continue
		}
		if string(c) == "\"" || string(c) == "'" {
			if i == len(input)-1 {
				continue
			}
			if !isString {
				isString = true

			} else {
				result = append(result, flag)
				flag = ""
				isFlagStarted = false
				isString = false
			}
			continue
		}
		if string(c) == " " && !isString {
			result = append(result, flag)
			flag = ""
			isFlagStarted = false
			continue
		}
		flag += string(c)
	}
	result = append(result, flag)
	return result
}

func removeDuplicateFlags(customFlags map[string]string, shortFlags map[string]string) {
	for longFlag, shortFlag := range shortFlags {
		if _, longExists := customFlags[longFlag]; longExists {
			if _, shortExists := customFlags[shortFlag]; shortExists {
				log.Entry().Warnf("Both forms of the flag %s and %s are presented, %s will be ignored.",
					longFlag, shortFlag, customFlags[shortFlag])
				delete(customFlags, shortFlag)
			}
		}
	}
}

func ParseCustomFlags(flagsStr string) map[string]string {
	flagsMap := make(map[string]string)
	parsedFlags := parseFlags(flagsStr)

	for _, flag := range parsedFlags {
		if strings.Contains(flag, "=") {
			split := strings.SplitN(flag, "=", 2)
			flagsMap[split[0]] = flag
		} else {
			flagsMap[flag] = flag
		}
	}

	removeDuplicateFlags(flagsMap, longShortFlagsMap)
	return flagsMap
}

func AppendThreadsAndRam(cmd []string, threads, ram string, customFlags map[string]string) []string {
	if len(threads) > 0 && !IsFlagSetByUser(customFlags, []string{"--threads", "-j"}) {
		cmd = append(cmd, "--threads="+threads)
	}
	if len(ram) > 0 && !IsFlagSetByUser(customFlags, []string{"--ram", "-M"}) {
		cmd = append(cmd, "--ram="+ram)
	}
	return cmd
}
