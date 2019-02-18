# mtaBuild

## Description

Executes the SAP Multitarget Application Archive Builder to create an mtar archive of the application.

Before doing this, validates that SAP Multitarget Application Archive Builder exists and the version is compatible.

Note that a version is formed by `major.minor.patch`, and a version is compatible to another version if the minor and patch versions are higher, but the major version is not, e.g. if 3.39.10 is the expected version, 3.39.11 and 3.40.1 would be compatible versions, but 4.0.1 would not be a compatible version.

## Parameters

| parameter        | mandatory | default                                                | possible values    |
| -----------------|-----------|--------------------------------------------------------|--------------------|
| `script`         | yes       |                                                        |                    |
| `dockerImage`    | no        | `ppiper/mta-archive-builder`                           |                    |
| `dockerOptions`  | no        | ''                                                     |                    |
| `buildTarget`    | yes       | `'NEO'`                                                | 'CF', 'NEO', 'XSA' |
| `extension`      | no        |                                                        |                    |
| `mtaJarLocation` | no        | `'/opt/sap/mta/lib/mta.jar'`                           |                    |
| `applicationName`| no        |                                                        |                    |

* `script` - The common script environment of the Jenkinsfile running. Typically the reference to the script calling the pipeline step is provided with the `this` parameter, as in `script: this`. This allows the function to access the [`commonPipelineEnvironment`](commonPipelineEnvironment.md) for retrieving, for example, configuration parameters.
* `dockerImage` - The Docker image to execute the MTA build.
  Note that you can provide your own image if required, but for most cases, the default should be fine.
* `dockerOptions` Docker options to be set when starting the container. It can be a list or a string.
* `buildTarget` - The target platform to which the mtar can be deployed.
* `extension` - The path to the extension descriptor file.
* `mtaJarLocation` - The location of the SAP Multitarget Application Archive Builder jar file, including file name and extension. First, the location is retrieved from the environment variables using the environment variable `MTA_JAR_LOCATION`. If no environment variable is provided, the location is retrieved from the parameters, or the step configuration using the key `mtaJarLocation`. If SAP Multitarget Application Archive Builder is not found on one of the previous locations an AbortException is thrown.
  Note that the environment variable `MTA_JAR_LOCATION` has priority. In case that the script runs on multiple nodes, SAP Multitarget Application Archive Builder must be located on all the nodes, therefore the environment variable must be also configured on all the nodes.
* `applicationName` - The name of the application which is being built. If the parameter has been provided and no `mta.yaml` exists, the `mta.yaml` will be automatically generated using this parameter and the information (`name` and `version`) from `package.json` before the actual build starts.

## Step configuration

The following parameters can also be specified as step parameters using the global configuration file:

* `dockerImage`
* `buildTarget`
* `extension`
* `mtaJarLocation`
* `applicationName`

## Side effects

1. The file name of the resulting archive is written to the `commonPipelineEnvironment` with variable name `mtarFileName`.

## Exceptions

* `AbortException`:
  * If SAP Multitarget Application Archive Builder is not found.
  * If there is an invalid `buildTarget`.
  * If there is no key `ID` inside the `mta.yaml` file.

## Example

```groovy
def mtarFileName
dir('/path/to/FioriApp'){
  mtarFileName = mtaBuild script:this, buildTarget: 'NEO'
}
```
