#!groovy

package util

import com.sap.piper.GitUtils
import com.sap.piper.Utils
import org.codehaus.groovy.runtime.InvokerHelper
import org.springframework.context.annotation.Bean
import org.springframework.context.annotation.Configuration

@Configuration
class BasePiperTestContext {

    @Bean
    Script nullScript() {
        def nullScript = InvokerHelper.createScript(null, new Binding())
        nullScript.currentBuild = [:]
        LibraryLoadingTestExecutionListener.prepareObjectInterceptors(nullScript)
        return nullScript
    }

    @Bean
    GitUtils mockGitUtils() {
        def mockGitUtils = new GitUtils()
        LibraryLoadingTestExecutionListener.prepareObjectInterceptors(mockGitUtils)
        return mockGitUtils
    }

    @Bean
    Utils mockUtils() {
        def mockUtils = new Utils()
        mockUtils.steps = [
            stash  : { m -> println "stash name = ${m.name}" },
            unstash: { println "unstash called ..." }
        ]
        LibraryLoadingTestExecutionListener.prepareObjectInterceptors(mockUtils)
        return mockUtils
    }
}
