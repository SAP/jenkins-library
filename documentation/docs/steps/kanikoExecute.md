# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

When pushing to a container registry, you need to maintain the respective credentials in your Jenkins credentials store:

Kaniko expects a Docker `config.json` file containing the credential information for registries.
You can create it like explained in the [protocodeExecuteScan Prerequisites section](https://www.project-piper.io/steps/protecodeExecuteScan/#prerequisites).

Please copy this file and upload it to your Jenkins for example<br />
via _Jenkins_ -> _Credentials_ -> _System_ -> _Global credentials (unrestricted)_ -> _Add Credentials_ ->

* Kind: _Secret file_
* File: upload your `config.json` file
* ID: specify id which you then use for the configuration of `dockerConfigJsonCredentialsId` (see below)

## ${docJenkinsPluginDependencies}

## Example

```groovy
kanikoExecute script:this
```

## Passing Environment Variables to Dockerfile

If you need to pass environment variables as build arguments to your Dockerfile (e.g., for `ARG` instructions), you must configure both `dockerOptions` and `buildOptions` in your pipeline configuration:

```yaml
kanikoExecute:
  dockerOptions:
    - --env BUILD_NUMBER
    - -u 0
    - --entrypoint=
  buildOptions:
    - --build-arg=BUILD_NUMBER
```

In your Dockerfile, you can then use the build argument:

```dockerfile
ARG BUILD_NUMBER
ENV BUILD_NUMBER=${BUILD_NUMBER:-dev}
RUN echo "BUILD_NUMBER during build is: $BUILD_NUMBER"
```

**Explanation:**

- `dockerOptions: --env BUILD_NUMBER` passes the environment variable from the CI/CD environment into the Kaniko container
- `dockerOptions: -u 0` runs as root user (required for Kaniko container execution)
- `dockerOptions: --entrypoint=` clears the default entrypoint (required for how Piper runs containers)
- `buildOptions: --build-arg=BUILD_NUMBER` passes the variable as a build argument to the Docker build process

Without both `--env` and `--build-arg` configured, the `ARG` in your Dockerfile will use its default value instead of the actual CI/CD environment variable.

## ${docGenParameters}

## ${docGenConfiguration}
