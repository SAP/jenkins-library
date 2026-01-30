# ${docGenStepName}

!!! warning "Deprecation notice"
This step will soon be deprecated!

## ${docGenDescription}

The SAP Alert Notification service for SAP BTP allows users to define
certain delivery channels, for example, e-mail or triggering of HTTP
requests, to receive notifications from pipeline events.

## Prerequisites

A service-key credential from the alert notification service.

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Example

Example configuration for the use in a `Jenkinsfile`.

```groovy
ansSendEvent(
  script: this,
  ansServiceKeyCredentialsId: "myANSCredential",
  eventType: "errorEvent",
  severity: "ERROR",
  category: "EXCEPTION",
  subject: "Something went wrong",
  body: "The details of what went wrong",
  priority: 3,
  tags: [
    myTag: "myValue",
    yourTag: "yourValue"
  ],
  resourceName: "Test Pipeline",
  resourceType: "My Pipeline",
  resourceInstance: "myPipeline",
  resourceTags: [
    myResourceTag: "a value"
  ]
)
```

Example for the use in a YAML configuration file (such as `.pipeline/config.yaml`).

```yaml
steps:
  <...>
  ansSendEvent:
    ansServiceKeyCredentialsId: "myANSCredential",
    eventType: "errorEvent",
    severity: "ERROR",
    category: "EXCEPTION",
    subject: "Something went wrong",
    body: "The details of what went wrong",
    priority: 3,
    tags:
      myTag: "myValue",
      yourTag: "yourValue",
    resourceName: "Test Pipeline",
    resourceType: "My Pipeline",
    resourceInstance: "myPipeline",
    resourceTags:
      myResourceTag: "a value"
```
