# Troubleshooting "Error response from daemon: toomanyrequests: You have reached your pull rate limit"

You may face the following error in your pipelines:

```
docker pull <some image>
Using default tag: latest
Error response from daemon: toomanyrequests: You have reached your pull rate limit. You may increase the limit by authenticating and upgrading: https://www.docker.com/increase-rate-limit
```

Those occur because Docker Hub has introduced rate limiting in November 2020. More background information is available [here](https://www.docker.com/pricing/resource-consumption-updates).

There are various options to mitigate this issue, which are listed below in no particular order.
No single option will work in *all* use-cases, please pick what works best for you.

## Company-internal Docker Hub mirror

If your company uses Artifactory for example, you might want to check if [Docker Hub mirroring](https://jfrog.com/knowledge-base/how-to-configure-a-remote-repository-in-artifactory-to-proxy-a-private-docker-registry-in-docker-hub/) is already enabled for you.

You could configure that registry for example using this snippet in your `.pipeline/config.yml` file.

```
steps:
  dockerExecute:
    dockerRegistryUrl: 'https://my.internal.registry:1234'
```

## Authenticated pulls from Docker Hub

The [`dockerExecute`](../steps/dockerExecute) step has an option `dockerRegistryCredentialsId` which you can use with any Docker Hub account.
See [Docker's information on pricing](https://www.docker.com/pricing) to check which type of account is right for you.

## Alternative Docker registry

Project "Piper"'s Docker images are also published to [GitHub Container Registry](https://github.com/orgs/SAP/packages?tab=packages&q=ppiper).
We don't have much experience with that, but in case the other options don't work for you, you might want to try consuming the images from there.

## Hyperscaler mirror

If you use some kind of hyperscaler, your provider might offer a Docker Hub mirror for you.
Please check the respective documentation of your provider.
