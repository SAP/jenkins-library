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
import static org.hamcrest.Matchers.contains
import static org.hamcrest.Matchers.hasItem
import static org.junit.Assert.assertEquals
import static org.junit.Assert.assertNull
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
    void inferBuildToolMaven() {
        nullScript.commonPipelineEnvironment.configuration = [
            general: [
                inferBuildTool: true
            ]
        ]
        helper.registerAllowedMethod('fileExists', [String.class], { s ->
            return s == "pom.xml"
        })
        def actual = nullScript.commonPipelineEnvironment.inferBuildTool(nullScript)
        assertEquals('maven', actual)
    }

    @Test
    void inferBuildToolMTA() {
        nullScript.commonPipelineEnvironment.configuration = [
            general: [
                inferBuildTool: true
            ]
        ]
        helper.registerAllowedMethod('fileExists', [String.class], { s ->
            return s == "mta.yaml"
        })
        def actual = nullScript.commonPipelineEnvironment.inferBuildTool(nullScript)
        assertEquals('mta', actual)
    }

    @Test
    void inferBuildToolNpm() {
        nullScript.commonPipelineEnvironment.configuration = [
            general: [
                inferBuildTool: true
            ]
        ]
        helper.registerAllowedMethod('fileExists', [String.class], { s ->
            return s == "package.json"
        })
        def actual = nullScript.commonPipelineEnvironment.inferBuildTool(nullScript)
        assertEquals('npm', actual)
    }

    @Test
    void inferBuildToolNone() {
        nullScript.commonPipelineEnvironment.configuration = [
            general: [
                inferBuildTool: true
            ]
        ]
        helper.registerAllowedMethod('fileExists', [String.class], { s ->
            return false
        })
        def actual = nullScript.commonPipelineEnvironment.inferBuildTool(nullScript)
        assertNull(actual)
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
        nullScript.commonPipelineEnvironment.setAbapRepositoryNames('[\"value1\",\"value2\"]')
        nullScript.commonPipelineEnvironment.writeToDisk(nullScript)


        assertThat(writeFileRule.files['.pipeline/commonPipelineEnvironment/artifactVersion'], is('1.0.0'))
        assertThat(writeFileRule.files['.pipeline/commonPipelineEnvironment/originalArtifactVersion'], is('2.0.0'))
        assertThat(writeFileRule.files['.pipeline/commonPipelineEnvironment/container/image'], is('myImage'))
        assertThat(writeFileRule.files['.pipeline/commonPipelineEnvironment/custom/custom1'], is('customVal1'))
        assertThat(writeFileRule.files['.pipeline/commonPipelineEnvironment/abap/repositoryNames'], is('[\"value1\",\"value2\"]'))
    }

    @Test
    void readFromDisk() {

        fileExistsRule.existingFiles.addAll([
            '.pipeline/commonPipelineEnvironment/artifactVersion',
            '.pipeline/commonPipelineEnvironment/originalArtifactVersion',
            '.pipeline/commonPipelineEnvironment/custom/custom1',
            '.pipeline/commonPipelineEnvironment/abap/repositoryNames',
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
            '.pipeline/commonPipelineEnvironment/abap/repositoryNames': '[\"value1\",\"value2\"]',
        ])

        nullScript.commonPipelineEnvironment.readFromDisk(nullScript)

        assertThat(nullScript.commonPipelineEnvironment.artifactVersion, is('1.0.0'))
        assertThat(nullScript.commonPipelineEnvironment.originalArtifactVersion, is('2.0.0'))
        assertThat(nullScript.commonPipelineEnvironment.valueMap['custom1'], is('customVal1'))
        assertThat(nullScript.commonPipelineEnvironment.abapRepositoryNames, is("[\"value1\",\"value2\"]"))
    }

}
