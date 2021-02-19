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
    void testContainereMap() {
        nullScript.commonPipelineEnvironment.setContainerProperty('image', 'myImage')
        assertThat(nullScript.commonPipelineEnvironment.getContainerProperty('image'), is('myImage'))
    }

    @Test
    void testWritetoDisk() {
        nullScript.commonPipelineEnvironment.artifactVersion = '1.0.0'
        nullScript.commonPipelineEnvironment.originalArtifactVersion = '2.0.0'
        nullScript.commonPipelineEnvironment.setContainerProperty('image', 'myImage')
        nullScript.commonPipelineEnvironment.setValue('custom1', 'customVal1')
        nullScript.commonPipelineEnvironment.setAbapAddonDescriptor('[\"value1\",\"value2\"]')
        nullScript.commonPipelineEnvironment.writeToDisk(nullScript)


        assertThat(writeFileRule.files['.pipeline/commonPipelineEnvironment/artifactVersion'], is('1.0.0'))
        assertThat(writeFileRule.files['.pipeline/commonPipelineEnvironment/originalArtifactVersion'], is('2.0.0'))
        assertThat(writeFileRule.files['.pipeline/commonPipelineEnvironment/container/image'], is('myImage'))
        assertThat(writeFileRule.files['.pipeline/commonPipelineEnvironment/custom/custom1'], is('customVal1'))
        assertThat(writeFileRule.files['.pipeline/commonPipelineEnvironment/abap/addonDescriptor'], is('[\"value1\",\"value2\"]'))
    }

    @Test
    void readFromDisk() {

        fileExistsRule.existingFiles.addAll([
            '.pipeline/commonPipelineEnvironment/artifactVersion',
            '.pipeline/commonPipelineEnvironment/originalArtifactVersion',
            '.pipeline/commonPipelineEnvironment/custom/custom1',
            '.pipeline/commonPipelineEnvironment/abap/addonDescriptor',
        ])

        nullScript.metaClass.findFiles {
            [
                [
                    'getName': {'custom1'},
                    'getPath': {'.pipeline/commonPipelineEnvironment/custom'},
                ]
            ]
        }

        readFileRule.files.putAll([
            '.pipeline/commonPipelineEnvironment/artifactVersion': '1.0.0',
            '.pipeline/commonPipelineEnvironment/originalArtifactVersion': '2.0.0',
            '.pipeline/commonPipelineEnvironment/custom': 'customVal1',
            '.pipeline/commonPipelineEnvironment/abap/addonDescriptor': '[\"value1\",\"value2\"]',
        ])

        nullScript.commonPipelineEnvironment.readFromDisk(nullScript)

        assertThat(nullScript.commonPipelineEnvironment.artifactVersion, is('1.0.0'))
        assertThat(nullScript.commonPipelineEnvironment.originalArtifactVersion, is('2.0.0'))
        assertThat(nullScript.commonPipelineEnvironment.valueMap['custom1'], is('customVal1'))
        assertThat(nullScript.commonPipelineEnvironment.abapAddonDescriptor, is("[\"value1\",\"value2\"]"))
    }

    @Test
    void writeAndReadFromDisk() {
        nullScript.commonPipelineEnvironment.setValue('string', 'testString')
        nullScript.commonPipelineEnvironment.setValue('boolean', true)
        nullScript.commonPipelineEnvironment.setValue('integer', 1)
        nullScript.commonPipelineEnvironment.setValue('list', ['item1', 'item2'])
        nullScript.commonPipelineEnvironment.setValue('map', [key1: 'val1', key2: 'val2'])

        nullScript.commonPipelineEnvironment.writeToDisk(nullScript)
        nullScript.commonPipelineEnvironment.readFromDisk(nullScript)

        assertThat(nullScript.commonPipelineEnvironment.getValue('string'), is('testString'))
        assertThat(nullScript.commonPipelineEnvironment.getValue('boolean'), is(true))
        assertThat(nullScript.commonPipelineEnvironment.getValue('integer'), is(1))
        assertThat(nullScript.commonPipelineEnvironment.getValue('list'), is(['item1', 'item2']))
        assertThat(nullScript.commonPipelineEnvironment.getValue('map'), is([key1: 'val1', key2: 'val2']))
    }
}
