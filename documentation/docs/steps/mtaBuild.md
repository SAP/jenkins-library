# mtaBuild

## Description
Executes the SAP MTA Archive Builder to create an mtar archive of the application.

## Prerequisites

* **SAP MTA Archive Builder** - available for download on the SAP Marketplace.
* **Java 8 or higher** - necessary to run the `mta.jar` file.
* **NodeJS installed** - the MTA Builder uses `npm` to download node module dependencies such as `grunt`.

## Parameters

| parameter        | mandatory | default                           | possible values    |
| -----------------|-----------|-----------------------------------|--------------------|
| `script`         | yes       |                                   |                    |
| `buildTarget`    | yes       |                                   | 'CF', 'NEO', 'XSA' |
| `mtaJarLocation` | no        |                                   |                    |

* `script`  The common script environment of the Jenkinsfile running. Typically the reference to the script calling the pipeline step is provided with the `this` parameter, as in `script: this`. This allows the function to access the [`commonPipelineEnvironment`](commonPipelineEnvironment.md) for retrieving, for example, configuration parameters.
* `buildTarget` The target platform to which the mtar can be deployed.
* `mtaJarLocation` The path of the `mta.jar` file. If no parameter is provided, the path is retrieved from the Jenkins environment variables using `env.MTA_JAR_LOCATION`. If the Jenkins environment variable is not set it is assumed that `mta.jar` is located in the current working directory.

## Return value

The file name of the resulting archive is returned with this step. The file name is extracted from the key `ID` defined in `mta.yaml`.

## Side effects

1. The file name of the resulting archive is written to the `commonPipelineEnvironment` with variable name `mtarFileName`.
2. As version number the timestamp is written into the `mta.yaml` file, that is packaged into the built archive.

## Exceptions

* `AbortException`
    * If there is an invalid `buildTarget`.
    * If there is no key `ID` inside the `mta.yaml` file.

## Example
```groovy
def mtarFileName
dir('/path/to/FioriApp'){
  mtarFileName = mtaBuild script:this, buildTarget: 'NEO'
}
```
