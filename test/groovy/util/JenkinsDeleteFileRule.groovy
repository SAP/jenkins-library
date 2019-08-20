package util

import com.lesfurets.jenkins.unit.BasePipelineTest
import org.junit.rules.TestRule
import org.junit.runner.Description
import org.junit.runners.model.Statement
import org.yaml.snakeyaml.Yaml

class JenkinsDeleteFileRule implements TestRule {

    final BasePipelineTest testInstance
    /**
     * The list of file paths that should have been deleted.
     * Can be used in tests to assert that productive code
     * removed the correct files.
     */
    List<String> deletedFiles = new ArrayList<>()

    private Boolean mockDeletion = true

    JenkinsDeleteFileRule(BasePipelineTest testInstance) {
        this.testInstance = testInstance
    }

    /**
     * Allows you to skip the actual deletion for test purposes.
     * @param skip - if `true` will skip the actual deletion.
     * @return this rule.
     */
    JenkinsDeleteFileRule mockDeletion(Boolean mockDeletion) {
        this.mockDeletion = mockDeletion
        return this
    }

    @Override
    Statement apply(Statement base, Description description) {
        return statement(base)
    }

    private Statement statement(final Statement base) {
        return new Statement() {
            @Override
            void evaluate() throws Throwable {
                testInstance.helper.registerAllowedMethod("deleteFile", [Map], { Map m ->
                    String path = m?.path
                    if(!path) {
                        error "[DeleteFile] File path must not be null or empty."
                    }

                    deletedFiles.add(path)
                    deleteFile(path, mockDeletion)
                })

                base.evaluate()
            }
        }
    }

    /**
     * Deleting the file, skipping deletion if specified.
     * @param filePath - the path of the file to delete.
     * @param mockDeletion - if `true` will skip actual deletion.
     */
    private void deleteFile(String filePath, Boolean mockDeletion) {
        File originalFile = new File(filePath)
        if (originalFile.exists()) {

            if(mockDeletion) {
                return
            }

            boolean deleted = originalFile.delete()
            if (!deleted) {
                throw new RuntimeException("Could not delete file at ${filePath}!")
            }
        }
    }
}
