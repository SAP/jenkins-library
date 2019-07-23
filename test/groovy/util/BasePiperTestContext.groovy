package util

import com.sap.piper.DescriptorUtils
import com.sap.piper.GitUtils
import com.sap.piper.JenkinsUtils
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
        nullScript.env = [:]
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
            stash  : {  },
            unstash: {  }
        ]
        LibraryLoadingTestExecutionListener.prepareObjectInterceptors(mockUtils)
        return mockUtils
    }

    @Bean
    JenkinsUtils mockJenkinsUtils() {
        def mockJenkinsUtils = new JenkinsUtils()
        LibraryLoadingTestExecutionListener.prepareObjectInterceptors(mockJenkinsUtils)
        return mockJenkinsUtils
    }

    @Bean
    DescriptorUtils mockDescriptorUtils() {
        def mockDescriptorUtils = new DescriptorUtils()
        LibraryLoadingTestExecutionListener.prepareObjectInterceptors(mockDescriptorUtils)
        return mockDescriptorUtils
    }
}
