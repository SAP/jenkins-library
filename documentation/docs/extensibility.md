# Extensibility

There are several possibilities for extensibility besides the **[very powerful configuration](configuration.md)**:

## 1. Stage Exits

  You have to create a file like `<StageName>.groovy` for example `Acceptance.groovy` and store it in folder `.pipeline/extensions/` in your source code repository.

!!! note "Cloud SDK Pipeline"
    If you use the Cloud SDK Pipeline, the folder is named `pipeline/extensions/` (without the dot). For more information, please refer to [the Cloud SDK Pipeline documentation](https://github.com/SAP/cloud-s4-sdk-pipeline/blob/master/doc/pipeline/extensibility.md).

  The pipeline template checks if such a file exists and executes it, if present.
  A parameter that contains the following keys is passed to the extension:

  * `script`: defines the global script environment of the Jenkinsfile run. This makes sure that the correct configuration environment can be passed to project "Piper" steps and also allows access to for example the `commonPipelineEnvironment`.
  * `originalStage`: this will allow you to execute the "original" stage at any place in your script. If omitting a call to `originalStage()` only your code will be executed instead.
  * `stageName`: name of the current stage
  * `config`: configuration of the stage (including all defaults)

  Here a simple example for such an extension:

  ``` groovy
  void call(Map params) {
      //access stage name
      echo "Start - Extension for stage: ${params.stageName}"

      //access config
      echo "Current stage config: ${params.config}"

      //execute original stage as defined in the template
      params.originalStage()

      //access overall pipeline script object
      echo "Branch: ${params.script.commonPipelineEnvironment.gitBranch}"

      echo "End - Extension for stage: ${params.stageName}"
  }
  return this
  ```

!!! note "Init stage cannot be extended"
    Please note, the `Init` stage among other things also checks out your current repository.<br />Thus it is not possible to use extensions on this stage.

## 2. Central Custom Template

If you have multiple projects where you want to use a custom template, you could implement this similarly to [piperPipeline](https://github.com/SAP/jenkins-library/blob/master/vars/piperPipeline.groovy).

!!! note "How to not get decoupled"
    Typically, providing a custom template decouples you from centrally provided updates to your template including the stages.<br />
    Where applicable, you can re-use the stage implementations. This means, you will call e.g. `piperPipelineStageBuild()` as you can see in [piperPipeline](https://github.com/SAP/jenkins-library/blob/master/vars/piperPipeline.groovy).

    Using this approach you can at least benefit from innovations done in individual project "Piper" stages.

!!! note "When to go with a custom template"
    If the configuration possibilities are not sufficient for you and if _1. Stage Exits_ is not applicable.

## 3. Custom Jenkinsfile

Since project "Piper" fully builds on [Jenkins Pipelines as Code](https://jenkins.io/doc/book/pipeline-as-code/), you can also go with your complete custom `Jenkinsfile`.

!!! warning "Decoupling"
    If you go this route you will be decoupled from the innovations provided with project "Piper", unless you re-use for example stages (as indicated above under _2. Central Custom Templates_).

    **We recommend to use this only as last option for extensibility.**


## Further tips and information

When you consider to add additional capabilities your first stop should be the [Jenkins Pipeline Steps Reference](https://jenkins.io/doc/pipeline/steps/).
Here you get an overview about what kind of capabilities are already available and a list of related parameters which you can use to customize the existing implementation. The provided information should help you to understand and extend the functionality of your pipeline.

!!! tip
    If you consider extensions we recommend you to do it using a custom library according to the [Jenkins shared libraries](https://jenkins.io/doc/book/pipeline/shared-libraries/) concept instead of adding groovy coding to the `Jenkinsfile`.
    Your custom library can easily live next to the provided pipeline library.

    Your Jenkinsfile would then start like

    ```
    @Library(['piper-lib-os', 'your-custom-lib']) _

    ```

<!-- ## Examples

work in progress
-->
