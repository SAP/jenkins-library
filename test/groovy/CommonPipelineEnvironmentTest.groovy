import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.JenkinsFileExistsRule
import util.JenkinsReadFileRule
import util.JenkinsWriteFileRule
import util.JenkinsReadYamlRule
import util.Rules

import static org.hamcrest.CoreMatchers.is
import static org.hamcrest.Matchers.hasItem
import static org.junit.Assert.assertThat

import org.junit.After

class CommonPipelineEnvironmentTest extends BasePiperTest {

    private JenkinsWriteFileRule writeFileRule = new JenkinsWriteFileRule(this)
    private JenkinsFileExistsRule fileExistsRule = new JenkinsFileExistsRule(this, [])
    private JenkinsReadFileRule readFileRule = new JenkinsReadFileRule(this, null)

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(writeFileRule)
        .around(fileExistsRule)
        .around(readFileRule)

    @After
    void tearDown() {
        nullScript.metaClass.findFiles = null
    }

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
    void testKeyRemoveFromMap() {
        nullScript.commonPipelineEnvironment.setValue('myList', [])
        nullScript.commonPipelineEnvironment.getValue('myList').add('item1')
        nullScript.commonPipelineEnvironment.removeValue('myList')
        assertThat(nullScript.commonPipelineEnvironment.getValue('myList'), is(null))
    }

    @Test
    void testContainereMap() {
        nullScript.commonPipelineEnvironment.setContainerProperty('image', 'myImage')
        assertThat(nullScript.commonPipelineEnvironment.getContainerProperty('image'), is('myImage'))
    }

}
