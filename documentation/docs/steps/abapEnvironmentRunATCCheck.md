# ${docGenStepName}

## ${docGenDescription}

!!! Currently the Object Set configuration is limited to the usage of Multi Property Sets. Please note that other sets besides the Multi Property Set will not be included in the ATC runs. You can see an example of the Multi Property Sets with all configurable properties. However, we strongly reccommend to only specify packages and software components like in the first two examples of the section `ATC config file example`.

## Prerequisites

* A SAP BTP, ABAP environment system is available. On this system, a [Communication User](https://help.sap.com/viewer/65de2977205c403bbc107264b8eccf4b/Cloud/en-US/0377adea0401467f939827242c1f4014.html), a [Communication System](https://help.sap.com/viewer/65de2977205c403bbc107264b8eccf4b/Cloud/en-US/1bfe32ae08074b7186e375ab425fb114.html) and a [Communication Arrangement](https://help.sap.com/viewer/65de2977205c403bbc107264b8eccf4b/Cloud/en-US/a0771f6765f54e1c8193ad8582a32edb.html) is setup for the Communication Scenario “ABAP Test Cockpit - Test Integration (SAP_COM_0901)“. This can be done manually through the respective applications on the SAP BTP, ABAP environment system or through creating a service key for the system on Cloud Foundry with the parameters {“scenario_id”: “SAP_COM_0901", “type”: “basic”}. In a pipeline, you can do this with the step [cloudFoundryCreateServiceKey](https://sap.github.io/jenkins-library/steps/cloudFoundryCreateServiceKey/).
* You can either provide the ABAP endpoint configuration to directly trigger an ATC run on the ABAP system or optionally provide the Cloud Foundry parameters with your credentials to read a Service Key of a SAP BTP, ABAP environment system in Cloud Foundry that contains all the details of the ABAP endpoint to trigger an ATC run.
* Regardless if you chose an ABAP endpoint directly or reading a Cloud Foundry Service Key, you have to provide the configuration of the packages and software components you want to be checked in an ATC run in a .yml or .yaml file. This file must be stored in the same folder as the Jenkinsfile defining the pipeline.
* The software components and/or packages you want to be checked must be present in the configured system in order to run the check. Please make sure that you have created or pulled the respective software components and/or Packages in the SAP BTP, ABAP environment system.

Examples will be listed below.

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Examples

### Configuration in the config.yml

The recommended way to configure your pipeline is via the config.yml file. In this case, calling the step in the Jenkinsfile is reduced to one line:

```groovy
abapEnvironmentRunATCCheck script: this
```

If you want to provide the host and credentials of the Communication Arrangement directly, the configuration could look as follows:

```yaml
steps:
  abapEnvironmentRunATCCheck:
    abapCredentialsId: 'abapCredentialsId',
    host: 'https://myABAPendpoint.com',
    atcConfig: 'atcconfig.yml',
```

### ATC run via Cloud Foundry Service Key example in Jenkinsfile

The following example triggers an ATC run via reading the Service Key of an ABAP instance in Cloud Foundry.

You can store the credentials in Jenkins and use the cfCredentialsId parameter to authenticate to Cloud Foundry.
The username and password to authenticate to ABAP system will then be read from the Cloud Foundry service key that is bound to the ABAP instance.

This can be done accordingly:

```groovy
abapEnvironmentRunATCCheck(
    cfApiEndpoint : 'https://test.server.com',
    cfOrg : 'cfOrg',
    cfSpace: 'cfSpace',
    cfServiceInstance: 'myServiceInstance',
    cfServiceKeyName: 'myServiceKey',
    abapCredentialsId: 'cfCredentialsId',
    atcConfig: 'atcconfig.yml',
    script: this,
)
```

To trigger the ATC run an ATC config file `atcconfig.yml` will be needed. Check section 'ATC config file example' for more information.

### ATC run via direct ABAP endpoint configuration in Jenkinsfile

This  example triggers an ATC run directly on the ABAP endpoint.

In order to trigger the ATC run you have to pass the username and password for authentication to the ABAP endpoint via parameters as well as the ABAP endpoint/host. You can store the credentials in Jenkins and use the abapCredentialsId parameter to authenticate to the ABAP endpoint/host.

This must be configured as following:

```groovy
abapEnvironmentRunATCCheck(
    abapCredentialsId: 'abapCredentialsId',
    host: 'https://myABAPendpoint.com',
    atcConfig: 'atcconfig.yml',
    script: this,
)
```

To trigger the ATC run an ATC config file `atcconfig.yml` will be needed. Check section 'ATC config file example' for more information.

### ATC config file example

Providing a specifc ATC configuration is optional. If you are using a `repositories.yml` file for the `Clone` stage of the ABAP environment pipeline, a default ATC configuration will be derived if no explicit ATC configuration is available.

The following section contains an example of an `atcconfig.yml` file.
This file must be stored in the same Git folder where the `Jenkinsfile` is stored to run the pipeline. This folder must be taken as a SCM in the Jenkins pipeline to run the pipeline.

You can specify a list of packages and/or software components to be checked. This must be in the same format as below example for a `atcconfig.yml` file.
In case subpackages shall be included in the checks you can use packagetrees.
Please note that if you chose to provide both packages and software components to be checked with the `atcconfig.yml` file, the set of packages and the set of software components will be combinend by the API using a logical AND operation.
Therefore, we advise to specify either the software components or packages.
Additionally, if you don't specify a dedicated ATC check variant to be used, the `ABAP_CLOUD_DEVELOPMENT_DEFAULT` variant will be used as default. For more information on how to configure a check variant for an ATC run please check the last example on this page.

See below example for an `atcconfig.yml` file with both packages and software components to be checked:

```yaml
objectset:
  softwarecomponents:
    - name: TestComponent
    - name: TestComponent2  
  packages:
    - name: TestPackage
  packagetrees:
    - name: TestPackageWithSubpackages
```

The following example of an `atcconfig.yml` file that only contains packages and packagetrees to be checked:

```yaml
objectset:
  packages:
    - name: TestPackage
  packagetrees:
    - name: TestPackageWithSubpackages
```

The following example of an `atcconfig.yml` file that only contains software components to be checked:

```yaml
objectset:
  softwarecomponents:
    - name: TestComponent
    - name: TestComponent2
```

The following is an example of an `atcconfig.yml` file that supports the check variant and configuration ATC options and containing the software components `TestComponent` and `TestComponent2` as Objectset.

```yaml
checkvariant: "TestCheckVariant"
configuration: "TestConfiguration"
objectset:
  softwarecomponents:
    - name: TestComponent
    - name: TestComponent2
```

The following example of an `atcconfig.yml` file contains all possible properties of the Multi Property Set that can be used. Please take note that this is not the reccommended approach. If you want to check packages or software components please use the two above examples. The usage of the Multi Property Set is only reccommended for ATC runs that require these rules for the test execution. There is no official documentation on the usage of the Multi Property Set.

```yaml
checkvariant: "TestCheckVariant"
configuration: "TestConfiguration"
objectset:
  type: multiPropertySet
  multipropertyset:
    owners:
      - name: demoOwner
    softwarecomponents:
      - name: demoSoftwareComponent
    versions:
      - value: ACTIVE
    packages:
      - name: demoPackage
    packagetrees:
      - name: TestPackageWithSubpackages
    objectnamepatterns:
      - value: 'ZCL_*'
    languages:
      - value: EN
    sourcesystems:
      - name: H01
    objecttypes:
      - name: CLAS
    objecttypegroups:
      - name: CLAS
    releasestates:
      - value: RELEASED
    applicationcomponents:
      - name: demoApplicationComponent
    transportlayers:
      - name: H01
```
