# Example Configurations

This page shows you some pipeline configuration examples.

As `Jenkinsfile` only following code is required:

```
@Library('piper-lib') _

piperPipeline script: this
```

## Pure Pull-Request Voting

.pipeline/config.yml:

``` YAML
general:
  buildTool: 'npm'
```

## Using custom defaults

It is possible to use custom defaults as indicated on the section about [Configuration](../configuration.md).

In order to use a custom defaults only a simple extension to the `Jenkinsfile` is required:

```
@Library(['piper-lib-os', 'myCustomLibrary']) _

piperPipeline script: this, customDefaults: ['myCustomDefaults.yml']
```

## more examples to come
