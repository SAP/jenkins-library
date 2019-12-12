# JSON Schema

The `metadata.json` file is a JSON schema for the step metadata located in [resource/metadata](../metadata).

## Usage

The file can be used with any YAML schema validator.

### VSCode

To use the schema in VSCode, install the [vscode-yaml](https://marketplace.visualstudio.com/items?itemName=redhat.vscode-yaml) extension to gain schema support for your `yaml` files.
Add the following code to your `.vscode/settings.json` file in the jenkins-library project:

```json
{
    "yaml.schemas": {
        "./resources/schemas/metadata.json": "resources/metadata/*.yaml"
    }
}
```
