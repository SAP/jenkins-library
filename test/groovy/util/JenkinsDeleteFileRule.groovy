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

    private Boolean skip

    JenkinsDeleteFileRule(BasePipelineTest testInstance) {
        this.testInstance = testInstance
    }

    /**
     * Allows you to skip the actual deletion for test purposes.
     * @param skip - if `true` will skip the actual deletion.
     * @return this rule.
     */
    JenkinsDeleteFileRule skipDeletion(Boolean skip) {
        this.skip = skip
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
                    deleteFile(path, skip)
                })

                base.evaluate()
            }
        }
    }

    /**
     * Deleting the file, skipping deletion if specified.
     * @param filePath - the path of the file to delete.
     * @param skipDeletion - if `true` will skip actual deletion.
     */
    private void deleteFile(String filePath, Boolean skipDeletion) {
        File originalFile = new File(filePath)
        if (originalFile.exists()) {

            if(skipDeletion) {
                return
            }

            boolean deleted = originalFile.delete()
            if (!deleted) {
                throw new RuntimeException("Could not delete file at ${filePath}!")
            }
        }
    }
}
