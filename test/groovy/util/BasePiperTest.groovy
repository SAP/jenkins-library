package util

import com.lesfurets.jenkins.unit.BasePipelineTest
import com.sap.piper.GitUtils
import com.sap.piper.JenkinsUtils
import com.sap.piper.Utils
import org.junit.Before
import org.junit.runner.RunWith
import org.springframework.beans.factory.annotation.Autowired
import org.springframework.test.annotation.DirtiesContext
import org.springframework.test.context.ContextConfiguration
import org.springframework.test.context.TestExecutionListeners
import org.springframework.test.context.junit4.SpringJUnit4ClassRunner

@RunWith(SpringJUnit4ClassRunner)
@ContextConfiguration(classes = [BasePiperTestContext.class])
@DirtiesContext(classMode = DirtiesContext.ClassMode.AFTER_EACH_TEST_METHOD)
@TestExecutionListeners(listeners = [LibraryLoadingTestExecutionListener.class], mergeMode = TestExecutionListeners.MergeMode.MERGE_WITH_DEFAULTS)
abstract class BasePiperTest extends BasePipelineTest {

    @Autowired
    Script nullScript

    @Autowired
    Utils utils

    @Autowired
    JenkinsUtils jenkinsUtils

    @Override
    @Before
    void setUp() throws Exception {
        helper = LibraryLoadingTestExecutionListener.singletonInstance
        if(!isHelperInitialized()) {
            super.setScriptRoots('.', 'vars')
            super.setUp()
        }
    }

    boolean isHelperInitialized() {
        try {
            helper.loadScript('dummy.groovy')
        } catch (Exception e) {
            if (e.getMessage().startsWith('GroovyScriptEngine is not initialized:'))
                return false
        }
        return true
    }

    @Deprecated
    void prepareObjectInterceptors(Object object) {
        LibraryLoadingTestExecutionListener.prepareObjectInterceptors(object)
    }
}
