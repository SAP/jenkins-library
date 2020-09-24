import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.JenkinsFileExistsRule
import util.JenkinsReadFileRule
import util.JenkinsReadYamlRule
import util.JenkinsStepRule
import util.JenkinsWriteFileRule
import util.Rules

import static org.junit.Assert.assertEquals
import static org.junit.Assert.assertFalse
import static org.junit.Assert.assertNull
import static org.junit.Assert.assertTrue

class PiperLoadGlobalExtensionsTest extends BasePiperTest {

    private Map checkoutParameters
    private boolean checkoutCalled = false
    private List filesRead = []
    private List fileWritten = []

    private JenkinsStepRule stepRule = new JenkinsStepRule(this)
    private JenkinsReadYamlRule readYamlRule = new JenkinsReadYamlRule(this)
    private JenkinsFileExistsRule fileExistsRule = new JenkinsFileExistsRule(this, [])

    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(stepRule)
        .around(readYamlRule)
        .around(fileExistsRule)

    @Before
    void init() {
        helper.registerAllowedMethod("checkout", [Map.class], { map ->
            checkoutParameters = map
            checkoutCalled = true
        })
        helper.registerAllowedMethod("readFile", [Map.class], { map ->
            filesRead.add(map.file)
            return ""
        })
        helper.registerAllowedMethod("writeFile", [Map.class], { map ->
            fileWritten.add(map.file)
        })
    }

    @Test
    void testNotConfigured() throws Exception {
        stepRule.step.piperLoadGlobalExtensions(script: nullScript)
        assertFalse(checkoutCalled)
    }

    @Test
    void testUrlConfigured() throws Exception {

        nullScript.commonPipelineEnvironment.configuration = [
            general: [
                globalExtensionsRepository: 'https://my.git.example/foo/bar.git'
            ]
        ]

        stepRule.step.piperLoadGlobalExtensions(script: nullScript)
        assertTrue(checkoutCalled)
        assertEquals('GitSCM', checkoutParameters.$class)
        assertEquals(1, checkoutParameters.userRemoteConfigs.size())
        assertEquals([url: 'https://my.git.example/foo/bar.git'], checkoutParameters.userRemoteConfigs[0])
    }

    @Test
    void testVersionConfigured() throws Exception {

        nullScript.commonPipelineEnvironment.configuration = [
            general: [
                globalExtensionsRepository: 'https://my.git.example/foo/bar.git',
                globalExtensionsVersion: 'v35'
            ]
        ]

        stepRule.step.piperLoadGlobalExtensions(script: nullScript)
        assertTrue(checkoutCalled)
        assertEquals(1, checkoutParameters.branches.size())
        assertEquals([name: 'v35'], checkoutParameters.branches[0])
    }

    @Test
    void testCredentialsConfigured() throws Exception {

        nullScript.commonPipelineEnvironment.configuration = [
            general: [
                globalExtensionsRepository: 'https://my.git.example/foo/bar.git',
                globalExtensionsRepositoryCredentialsId: 'my-credentials'
            ]
        ]

        stepRule.step.piperLoadGlobalExtensions(script: nullScript)
        assertTrue(checkoutCalled)
        assertEquals(1, checkoutParameters.userRemoteConfigs.size())
        assertEquals([url: 'https://my.git.example/foo/bar.git', credentialsId: 'my-credentials'], checkoutParameters.userRemoteConfigs[0])
    }

    @Test
    void testExtensionConfigurationExists() throws Exception {
        fileExistsRule.registerExistingFile('test/extension_configuration.yml')

        nullScript.commonPipelineEnvironment.configuration = [
            general: [
                globalExtensionsDirectory: 'test',
                globalExtensionsRepository: 'https://my.git.example/foo/bar.git'
            ]
        ]

        Map prepareParameter = [:]
        helper.registerAllowedMethod("prepareDefaultValues", [Map.class], { map ->
            prepareParameter = map
        })

        stepRule.step.piperLoadGlobalExtensions(script: nullScript, customDefaults: ['default.yml'], customDefaultsFromFiles: ['file1.yml'])
        assertTrue(checkoutCalled)

        //File copied
        assertTrue(filesRead.contains('test/extension_configuration.yml'))
        assertTrue(fileWritten.contains('.pipeline/extension_configuration.yml'))

        assertEquals(2, prepareParameter.customDefaultsFromFiles.size())
        assertEquals('extension_configuration.yml', prepareParameter.customDefaultsFromFiles[0])
        assertEquals('file1.yml', prepareParameter.customDefaultsFromFiles[1])
        assertEquals(1, prepareParameter.customDefaults.size())
        assertEquals('default.yml', prepareParameter.customDefaults[0])
    }

    @Test
    void testLoadLibraries() throws Exception {
        fileExistsRule.registerExistingFile('test/sharedLibraries.yml')

        nullScript.commonPipelineEnvironment.configuration = [
            general: [
                globalExtensionsDirectory: 'test',
                globalExtensionsRepository: 'https://my.git.example/foo/bar.git'
            ]
        ]

        readYamlRule.registerYaml("test/sharedLibraries.yml", "[{name: my-extension-dependency, version: my-git-tag}]")

        List libsLoaded = []
        helper.registerAllowedMethod("library", [String.class], { lib ->
            libsLoaded.add(lib)
        })

        stepRule.step.piperLoadGlobalExtensions(script: nullScript)
        assertTrue(checkoutCalled)
        assertEquals(1, libsLoaded.size())
        assertEquals("my-extension-dependency@my-git-tag", libsLoaded[0].toString())
    }
}
