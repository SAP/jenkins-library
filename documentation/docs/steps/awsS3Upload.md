# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

* Before you can use the step awsS3Upload, you must have an Amazon account. See [How do I create and activate a new AWS account?](https://aws.amazon.com/premiumsupport/knowledge-center/create-and-activate-aws-account/) for details.
* You will need AWS access keys for your S3 Bucket. Access keys consist of an access key ID and secret access key, which are used to sign programmatic requests that you make to AWS. You can create them by using the AWS Management Console.
* The access keys must allow the action "s3:PutObject" for the specified S3 Bucket

## Set up the AWS Credentials

To make your AWS credentials available to the jenkins library, store them as Jenkins credentials of type "Secret Text". The "Secret Text" must be in JSON format and contain the "access_key_id", "secret_access_key", "bucket" as well as the "region".

For Example:

```JSON
{
  "access_key_id": "FJNAKNCLAVLRNBLAVVBK",
  "bucket": "vro-artloarj-ltnl-nnbv-ibnh-lbnlsnblltbn",
  "secret_access_key": "123467895896646438486316436kmdlcvreanvjk",
  "region": "eu-central-1"
}
```

If the JSON string contains additional information, this is not a problem. These are automatically detected and skipped.

## About Files/Directories to Upload

With the step awsS3Upload you can upload single files as well as whole directories into your S3 bucket. File formats do not matter and directory structures are preserved.

**Note:** File paths must be specified in UNIX format. So the used path separator must be "/".

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Example

```groovy
awsS3Upload(
    script: this,
    awsCredentialsId: "AWS_Credentials",
    filePath: "test.txt"
)
```
