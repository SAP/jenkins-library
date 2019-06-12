package util

import com.lesfurets.jenkins.unit.global.lib.SourceRetriever

/**
 * Retrieves the shared lib sources of the current project which are expected to be
 * at the default location &quot;./vars&quot;.
 */

class ProjectSource implements SourceRetriever {

    private def sourceDir = new File('.')

    /*
     * None of the parameters provided in the signature are used in the use-case of that retriever.
     */
    List<URL> retrieve(String repository, String branch, String targetPath) {
        if (sourceDir.exists()) {
            return [sourceDir.toURI().toURL()]
        }
        throw new IllegalStateException("Directory $sourceDir.path does not exists!")
    }
}
