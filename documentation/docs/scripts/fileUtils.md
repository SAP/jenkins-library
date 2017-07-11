# FileUtils

## Description
Provides file system related utility functions.

## Constructor
Since there are only static utility methods there is no need for instantiating objects. 

## Method Details

### validateDirectory(dir)

#### Description
Checks whether a file exists and is a directory.

#### Parameters

* `dir` - directory to be checked. In case it is relative path it is checked against the
current working directory. In case of doubt use the absolute path (prefix the directory with `pwd`).

#### Return value
none

#### Side effects
none

#### Exceptions
* `IllegalArgumentException`: If the parameter `dir` is null or empty.
* `AbortException`: If the directory does not exist or is not a directory.

#### Example

```groovy
FileUtils.validateDirectory('/path/to/dir')
```

### validateDirectoryIsNotEmpty(dir)

#### Description
Check whether a directory is not empty. Before the directory is checked, `validateDirectory(dir)` is executed.

#### Parameters

* `dir` - directory to be checked. In case it is relative path it is checked against the
current working directory. In case of doubt use the absolute path (prefix the directory with `pwd`).

#### Return value
none

#### Side effects
none

#### Exceptions
* `IllegalArgumentException`: If the parameter `dir` is null or empty.
* `AbortException`: If the directory does not exist or is not a directory or the directory is empty.

#### Example

```groovy
FileUtils.validateDirectoryIsNotEmpty('/path/to/dir')
```

