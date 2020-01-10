import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.JenkinsFileExistsRule
import util.JenkinsWriteFileRule
import util.JenkinsReadYamlRule
import util.Rules

import static org.hamcrest.CoreMatchers.is
import static org.hamcrest.Matchers.contains
import static org.hamcrest.Matchers.hasItem
import static org.junit.Assert.assertThat

class CommonPipelineEnvironmentTest extends BasePiperTest {

    private JenkinsWriteFileRule writeFileRule = new JenkinsWriteFileRule(this)
    private JenkinsFileExistsRule fileExistsRule = new JenkinsFileExistsRule(this, [])

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(writeFileRule)
        .around(fileExistsRule)

    @Test
    void testCustomValueList() {
        nullScript.commonPipelineEnvironment.setValue('myList', [])
        nullScript.commonPipelineEnvironment.getValue('myList').add('item1')
        nullScript.commonPipelineEnvironment.getValue('myList').add('item2')
        assertThat(nullScript.commonPipelineEnvironment.getValue('myList'), hasItem('item1'))
        assertThat(nullScript.commonPipelineEnvironment.getValue('myList'), hasItem('item2'))
    }

    @Test
    void testCustomValueMap() {
        nullScript.commonPipelineEnvironment.setValue('myList', [:])
        nullScript.commonPipelineEnvironment.getValue('myList').key1 = 'val1'
        nullScript.commonPipelineEnvironment.getValue('myList').key2 = 'val2'
        assertThat(nullScript.commonPipelineEnvironment.getValue('myList').key1, is('val1'))
        assertThat(nullScript.commonPipelineEnvironment.getValue('myList').key2, is('val2'))
    }

    @Test
    void testWritetoDisk() {
        nullScript.commonPipelineEnvironment.artifactVersion = '1.0.0'
        nullScript.commonPipelineEnvironment.setValue('custom1', 'customVal1')
        nullScript.commonPipelineEnvironment.writeToDisk(nullScript)

        assertThat(writeFileRule.files['.pipeline/piperEnvironment/artifactVersion'], is('1.0.0'))
        assertThat(writeFileRule.files['.pipeline/piperEnvironment/custom/custom1'], is('customVal1'))
    }
}
