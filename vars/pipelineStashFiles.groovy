import com.sap.piper.GenerateDocumentation
import groovy.transform.Field

@Field STEP_NAME = getClass().getName()

@Field Set GENERAL_CONFIG_KEYS = []

@Field Set STEP_CONFIG_KEYS = []

@Field Set PARAMETER_KEYS = [
    /**
     * Can be used to overwrite the default behavior of existing stashes as well as to define additional stashes.
     * This parameter handles the _includes_ and can be defined as a map of stash name and include patterns.
     * Include pattern has to be a string with comma separated patterns as per [Pipeline basic step `stash`](https://jenkins.io/doc/pipeline/steps/workflow-basic-steps/#stash-stash-some-files-to-be-used-later-in-the-build)
     */
    'stashIncludes',
    /**
     * Can be used to overwrite the default behavior of existing stashes as well as to define additional stashes.
     * This parameter handles the _excludes_ and can be defined as a map of stash name and exclude patterns.
     * Exclude pattern has to be a string with comma separated patterns as per [Pipeline basic step `stash`](https://jenkins.io/doc/pipeline/steps/workflow-basic-steps/#stash-stash-some-files-to-be-used-later-in-the-build)
     */
    'stashExcludes'
]

/**
 * This step stashes files that are needed in other build steps (on other nodes).
 */
@GenerateDocumentation
void call(Map parameters = [:], body) {
    handlePipelineStepErrors (stepName: 'pipelineStashFiles', stepParameters: parameters) {

        pipelineStashFilesBeforeBuild(parameters)
        body() //execute build
        pipelineStashFilesAfterBuild(parameters)
    }
}
