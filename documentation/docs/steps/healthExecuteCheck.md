# healthExecuteCheck

## Description

Calls the health endpoint url of the application.

The intention of the check is to verify that a suitable health endpoint is available. Such a health endpoint is required for operation purposes.

This check is used as a real-life test for your productive health endpoints.

!!! note "Check Depth"
    Typically, tools performing simple health checks are not too smart. Therefore it is important to choose an endpoint for checking wisely.

    This check therefore only checks if the application/service url returns `HTTP 200`.

    This is in line with health check capabilities of platforms which are used for example in load balancing scenarios. Here you can find an [example for Amazon AWS](http://docs.aws.amazon.com/elasticloadbalancing/latest/classic/elb-healthchecks.html).

## Prerequisites

Endpoint for health check is configured.

!!! warning
    The health endpoint needs to be available without authentication!

!!! tip
    If using Spring Boot framework, ideally the provided `/health` endpoint is used and extended by development. Further information can be found in the [Spring Boot documenation for Endpoints](http://docs.spring.io/spring-boot/docs/current/reference/html/production-ready-endpoints.html)

## Example

Pipeline step:

```groovy
healthExecuteCheck testServerUrl: 'https://testserver.com'
```

## Parameters

| parameter | mandatory | default | possible values |
| ----------|-----------|---------|-----------------|
|script|yes|||
|healthEndpoint|no|``||
|testServerUrl|no|||

Details:
* `script` defines the global script environment of the Jenkinsfile run. Typically `this` is passed to this parameter. This allows the function to access the [`commonPipelineEnvironment`](commonPipelineEnvironment.md) for storing the measured duration.
* Health check function is called providing full qualified `testServerUrl` (and optionally with `healthEndpoint` if endpoint is not the standard url) to the health check.
* In case response of the call is different than `HTTP 200 OK` the **health check fails and the pipeline stops**.

## Step configuration

We recommend to define values of step parameters via [config.yml file](../configuration.md).

In following sections the configuration is possible:

| parameter | general | step | stage |
| ----------|-----------|---------|-----------------|
|script||||
|healthEndpoint|X|X|X|
|testServerUrl|X|X|X|
