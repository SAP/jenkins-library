# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

This step is intended for use in **GitHub Actions**, where Docker + BuildKit are natively available on the runner.
It is not supported on Jenkins or Azure DevOps runners.

When pushing to a container registry, you need to provide credentials via one of these approaches:

* Pass `containerRegistryUser` and `containerRegistryPassword` parameters — the step will write a `config.json` automatically.
* Provide a pre-existing `config.json` via the `dockerConfigJSON` parameter.

## ${docJenkinsPluginDependencies}

## Example

### Building a single image

```yaml
steps:
  dockerBuild:
    containerImageName: myImage
    containerRegistryUrl: my.registry.example.com
    containerRegistryUser: myUser
    containerRegistryPassword: myPassword
```

### Building multiple images from sub-directories

`containerRegistryUrl`, `containerRegistryUser`, and `containerRegistryPassword` are required —
the multi-image build path will fail immediately if `containerRegistryUrl` is empty.

```yaml
steps:
  dockerBuild:
    containerImageName: myImage
    containerRegistryUrl: my.registry.example.com
    containerRegistryUser: myUser
    containerRegistryPassword: myPassword
    containerMultiImageBuild: true
```

With the following Dockerfiles present in the repository:

* `sub1/Dockerfile`
* `sub2/Dockerfile`

The following images will be built and pushed:

* `myImage-sub1`
* `myImage-sub2`

### Using registry mirrors

```yaml
steps:
  dockerBuild:
    containerImageName: myImage
    registryMirrors:
      - mirror.gcr.io
      - mycompany-docker-virtual.jfrog.io
```

### Enabling BOM creation

```yaml
steps:
  dockerBuild:
    containerImageName: myImage
    createBOM: true
```

### Opting in (replacing kanikoExecute)

By default `kanikoExecute` is active and `dockerBuild` is off. To switch:

```yaml
stages:
  Build:
    dockerBuild: true
    kanikoExecute: false
```

## ${docGenParameters}

## ${docGenConfiguration}
