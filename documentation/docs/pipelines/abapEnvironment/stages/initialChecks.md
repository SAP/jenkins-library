# Initial Checks

This stage runs preliminary checks for the `Build` stage.

## Steps

The following steps are executed in this stage:

- [abapAddonAssemblyKitCheckPV](../../../steps/abapAddonAssemblyKitCheckPV.md)
- [abapAddonAssemblyKitCheckCVs](../../../steps/abapAddonAssemblyKitCheckCVs.md)
- [abapAddonAssemblyKitReserveNextPackages](../../../steps/abapAddonAssemblyKitReserveNextPackages.md)

## Stage Parameters

There are no specifc stage parameters.

## Stage Activation

This stage will be active, if the stage configuration in the `config.yml` contains entries for the `Build` stage.

## Configuration Example

Have a look at the [Build stage](build.md).
