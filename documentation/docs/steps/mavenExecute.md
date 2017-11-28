# mavenExecute

## Description

Executes a closure inside a docker container with the specified docker image. 
The workspace is mounted into the docker image.
Proxy environment variables defined on the Jenkins machine are also available in the Docker container.

## Parameters

| parameter            | mandatory | default           | example values             |
| ---------------------|-----------|-------------------|----------------------------|
| `dockerImage`        | no        | 'maven:3.5-jdk-7' |                            |
| `globalSettingsFile` | no        |                   | 'local_folder/settings.xml'|
| `projectSettingsFile`| no        |                   |                            |
| `pomPath`            | no        |                   | 'local_folder/m2'          |
| `flags`              | no        |                   | '-o'                       |
| `goals`              | no        |                   | 'clean install'            |
| `m2Path`             | no        |                   | 'local_folder/m2'          |
| `defines`            | no        |                   | '-Dmaven.tests.skip=true'  |

* `dockerImage` Name of the docker image that should be used. If empty Docker is not used.
* `globalSettingsFile` Path or url to the mvn settings file that should be used as global settings file. 
* `projectSettingsFile` Path or url to the mvn settings file that should be used as project settings file.
* `pomPath` Path the the pom file that should be used.
* `flags` Flags to provide when running mvn.
* `goals` Maven goals that should be executed.
* `m2Path` Path to the location of the local repository that should be used.
* `defines` Additional properties.

## Global Configuration
The following parameters can also be specified using the global configuration file:
* `dockerImage`
* `globalSettingsFile`
* `projectSettingsFile`
* `pomPath`
* `m2Path`

## Exceptions

None

## Example

```groovy
mavenExecute script: this, goals: 'clean install'
```




