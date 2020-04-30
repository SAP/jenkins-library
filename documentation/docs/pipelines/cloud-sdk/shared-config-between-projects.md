# Share Configuration Between Projects

SAP Cloud SDK Pipeline does not require any programming on the application developer's end, as the pipeline is centrally developed and maintained.
The necessary configuration happens in the `.pipeline/config.yml` file in the root directory of the application's repository.

For projects that are composed of multiple repositories (microservices), it might be desired to share the common configuration.
To do that, create a YAML file which is accessible from your CI/CD environment and configure it in your project.
For example, the common configuration can be stored in a GitHub repository an accessed via the "raw" URL:

```yaml
general:
  sharedConfiguration: 'https://my.github.local/raw/someorg/shared-config/master/backend-service.yml'
```

It is important to ensure that the HTTP response body is proper YAML, as the pipeline will attempt to parse it.

Anonymous read access to the `shared-config` repository is required.

The shared config is merged with the project's `.pipeline/config.yml`.
Note that the project's config takes precedence, so you can override the shared configuration in your project's local configuration.
This might be useful to provide a default value that needs to be changed only in some projects.

If you have different types of projects, they might require different shared configuration.
For example, you might not require all projects to have a certain code check (like Checkmarx, SourceClear, Whitesource) active.
This can be achieved by having multiple YAML files in the _shared-config_ repository.
Configure the URL to the respective configuration file in the projects as described above.
