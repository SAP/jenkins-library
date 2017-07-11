# Version

## Description
Handles version numbers.

## Constructors

### Version(major, minor, patch)

#### Parameters

| parameter | mandatory | default | possible values |
| ----------|-----------|---------|-----------------|
| major     | yes       |         |                 |
| minor     | yes       |         |                 |
| patch     | no        | `-1`    |                 |

* `major` - the major version number.
* `minor` - the minor version number.
* `patch` - the patch version number.

#### Exceptions

* `IllegalArgumentException`: If the `major` or `minor` version number is less than `0`.

#### Example

```groovy
def toolVersion = new Version(1, 2, 3)
```

### Version(text)

#### Parameters

| parameter | mandatory | default | possible values |
| ----------|-----------|---------|-----------------|
| text      | yes       |         |                 |

* `text` - as an alternative to calling the constructor with `major`, `minor`, and `patch` version numbers, you can pass this as a String of format 'major.minor.patch'.

#### Exceptions

* `IllegalArgumentException`: If the `text` parameter is `null` or empty.
* `AbortException`: If the version `text` has an unexpected format.

#### Example

```groovy
def toolVersion = new Version('1.2.3')
```

## Method Details

### equals

#### Description
Indicates whether some other version instance is equal to this one. The two versions are considered equal when they have the same `major`, `minor` and `patch` version number.

#### Parameters

* `version` - the Version instance to compare to this Version instance.

#### Return value

`true` if `major`, `minor` and `patch` version numbers are equal to each other. Otherwise `false`.

#### Side effects

none

#### Exceptions

* `AbortException`:  If the parameter `version` is `null`.

#### Example

```groovy
assert new Version('1.2.3').equals(new Version('1.2.3'))
```

### isCompatibleVersion

#### Description
Checks whether a version is compatible. Two versions are compatible if the major version number is the same, while the minor and patch version number are the same or higher.

#### Parameters

* `version` - the Version instance to compare to this Version instance.

#### Return value

`true` if this Version instance is compatible to the other Version instance. Otherwise `false`.

#### Side effects

none

#### Exceptions

* `AbortException`: If the parameter `version` is `null`.

#### Example

```groovy
assert new Version('1.2.3').isCompatibleVersion(new Version('1.3.1'))
```

### isHigher

#### Description
Checks whether this Version instance is higher than the other Version instance.

#### Parameters

* `version` - the Version instance to compare to this Version instance.

#### Return value

`true` if this Version instance is higher than the other Version instance. Otherwise `false`.

#### Side effects

none

#### Exceptions

* `AbortException`: If the parameter `version` is `null`.

#### Example

```groovy
assert new Version('1.2.3').isHigher(new Version('1.1.6'))
```

### toString

#### Description
Print the version number in format '<major>.<minor>.<patch>'. If no patch version number exists the format is '<major>.<minor>'.

#### Parameters

none

#### Return value

A String consisting of `major`, `minor` and if available `patch`, separated by dots.

#### Side effects

none

#### Exceptions

none

#### Example

```groovy
assert "${new Version('1.2.3')}" == "1.2.3"
```
