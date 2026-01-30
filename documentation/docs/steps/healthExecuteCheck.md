# ${docGenStepName}

!!! warning "Deprecation notice"
This step will soon be deprecated!

## ${docGenDescription}

## Prerequisites

Endpoint for health check is configured.

!!! warning
    The health endpoint needs to be available without authentication!

!!! tip
    If using Spring Boot framework, ideally the provided `/health` endpoint is used and extended by development. Further information can be found in the [Spring Boot documentation for Endpoints](http://docs.spring.io/spring-boot/docs/current/reference/html/production-ready-endpoints.html)

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Example

Pipeline step:

```groovy
healthExecuteCheck testServerUrl: 'https://testserver.com'
```
