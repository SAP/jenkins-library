# Custom Error Patterns for Steps

This document explains how to configure custom error patterns for pipeline steps to improve error detection and provide better user feedback.

## Overview

Custom error patterns allow steps to recognize specific error conditions in their output and provide enhanced error messages or categorization.

## Configuration

### Step Metadata Configuration

Error patterns are defined in the step's metadata YAML file under the `errors` section:

```yaml
metadata:
  name: myStep
  description: "My custom step"
  errors:
    - pattern: "npm error code E401"
      message: "NPM authentication failed. Check your credentials or token."
      category: "authentication"
    - pattern: "npm error 404.*Not Found"
      message: "NPM package not found. Check package name and registry."
      category: "dependency"
    - pattern: "ENOTFOUND"
      message: "Network error. Check your connection and registry URL."
      category: "network"
```

### Schema Definition

Each error pattern consists of:

- **`pattern`** (required): Regular expression or substring to match against step output
- **`message`** (optional): Enhanced error message to show users
- **`category`** (required): Error category for automated handling

## Implementation Details

### Pattern Matching

The system supports two matching modes:

1. **Regular Expression Matching** (preferred): Patterns are first treated as regex patterns
2. **Substring Matching** (fallback): If regex compilation fails, falls back to substring matching

Examples:

```yaml
# Regex pattern with wildcard
- pattern: "npm error 404.*Not Found"

# Exact substring match
- pattern: "npm error code E401"

# Regex with anchoring
- pattern: "^ERROR:"
```

### Runtime Behavior

During step execution:

1. All log messages are checked against configured error patterns
2. When a pattern matches:
   - The enhanced message is stored for potential use
   - In GitHub Actions environment, appropriate logging commands are used
   - The error category can be used for automated handling

## Example: NPM Step

The `npmExecuteScripts` step demonstrates comprehensive error pattern usage:

```yaml
metadata:
  name: npmExecuteScripts
  errors:
    # Authentication errors
    - pattern: "npm error code E401"
      category: "authentication"
    - pattern: "npm error Incorrect or missing password"
      message: "NPM authentication failed. Your password or token is incorrect."
      category: "authentication"

    # Dependency errors
    - pattern: "npm error 404.*Not Found"
      message: "NPM package not found. Check package name and registry."
      category: "dependency"

    # Network errors
    - pattern: "npm error ENOTFOUND"
      message: "NPM registry not reachable. Check network connection and registry URL."
      category: "network"

    # Permission errors
    - pattern: "npm error EACCES"
      message: "NPM permission denied. Check file permissions or registry access rights."
      category: "permission"

    # PNPM specific errors
    - pattern: "ERR_PNPM_FETCH_401"
      message: "PNPM authentication failed. Check your credentials or token."
      category: "authentication"
```

## Integration with GitHub Actions

When running in GitHub Actions environment, error patterns leverage GitHub Actions logging commands:

- `::error::` for error messages
- `::warning::` for warning messages
- `::notice::` for informational messages

This provides better integration with GitHub Actions workflow logs and UI.

## Best Practices

1. **Use Specific Patterns**: Make patterns as specific as possible to avoid false matches
2. **Provide Helpful Messages**: Include actionable guidance in error messages
3. **Categorize Appropriately**: Use consistent categories for similar error types
4. **Test Patterns**: Validate regex patterns work correctly with expected error output
5. **Document Intent**: Add comments in metadata YAML to explain complex patterns

## Migration Guide

To add error patterns to existing steps:

1. Edit the step's metadata YAML file in `resources/metadata/`
2. Add the `errors` section with desired patterns
3. Regenerate the step code: `go generate`
4. Test the step with scenarios that trigger the error patterns
5. Verify enhanced error messages appear correctly
