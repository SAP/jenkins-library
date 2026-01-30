# ${docGenStepName}

!!! warning "Deprecation notice"
This step will soon be deprecated!

## ${docGenDescription}

## Prerequisites

* Before you can access Azure Storage, you will need an Azure subscription. If you do not have a subscription, create an [account](https://azure.microsoft.com/en-us/).
* This step currently only supports authentication via Shared Access Signature (SAS).
* You can generate a SAS token from the Azure Portal under [Create a service SAS](https://docs.microsoft.com/en-us/rest/api/storageservices/create-service-sas).
* The SAS token must allow the actions "Write" and "Create" for the specified Azure Blob Storage.

## Set up the Azure Credentials

To make your Azure credentials available to the jenkins library, store them as Jenkins credentials of type "Secret Text". The "Secret Text" must be in JSON format and contain the "account_name", "container_name", as well as the "sas_token".

For Example:

```JSON
{
  "account_name": "asdfg12345jhgfdwertz4et5",
  "container_name": "abcde-lkjhg-qwertzui-fghj-9876-1234-7594rbnsmncx-xyz",
  "sas_token": "sig=1234567890wertzuiopaYXCVBNMASDsdfghjkloi1234567890qwedf%1993-12-15opphehttpsqtgcshje1234-aqwe-1234-5678-t57894u875LH2%nv23"
}
```

If the JSON string contains additional information, this is not a problem. These are skipped.

## About Files/Directories to Upload

With this step you can upload single files as well as whole directories into your Azure Storage. File formats do not matter and directory structures are preserved.

**Note:** File paths must be specified in UNIX format. So the used path separator must be "/".

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Example

```groovy
azureBlobUpload(
    script: this,
    azureCredentialsId: "Azure_Credentials",
    filePath: "test.txt"
)
```
