import com.sap.piper.DefaultValueCache
import com.sap.piper.Utils
import com.sap.piper.GitUtils
import org.junit.After
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import org.yaml.snakeyaml.Yaml
import util.BasePiperTest
import util.JenkinsLoggingRule
import util.JenkinsReadFileRule
import util.JenkinsShellCallRule
import util.JenkinsStepRule
import util.JenkinsWriteFileRule
import util.Rules

import static org.hamcrest.Matchers.is
import static org.junit.Assert.assertEquals
import static org.junit.Assert.assertNotNull
import static org.junit.Assert.assertNull
import static org.junit.Assert.assertThat

class SetupCommonPipelineEnvironmentTest extends BasePiperTest {

    def usedConfigFile
    def pipelineAndTestStashIncludes
    def utilsMock

    private JenkinsStepRule stepRule = new JenkinsStepRule(this)
    private JenkinsWriteFileRule writeFileRule = new JenkinsWriteFileRule(this)
    private ExpectedException thrown = ExpectedException.none()
    private JenkinsShellCallRule shellRule = new JenkinsShellCallRule(this)
    private JenkinsReadFileRule readFileRule = new JenkinsReadFileRule(this, "./")
    private JenkinsLoggingRule loggingRule = new JenkinsLoggingRule(this)

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(stepRule)
        .around(writeFileRule)
        .around(thrown)
        .around(shellRule)
        .around(readFileRule)
        .around(loggingRule)

    @Before
    void init() {

        def examplePipelineConfig = new File('test/resources/test_pipeline_config.yml').text

        helper.registerAllowedMethod("libraryResource", [String], { fileName ->
            switch(fileName) {
                case 'default_pipeline_environment.yml': return "default: 'config'"
                case 'custom.yml': return "custom: 'myConfig'"
                case 'notFound.yml': throw new hudson.AbortException('No such library resource notFound could be found')
                default: return "the:'end'"
            }
        })

        helper.registerAllowedMethod("readYaml", [Map], { Map parameters ->
            Yaml yamlParser = new Yaml()
            if (parameters.text) {
                return yamlParser.load(parameters.text)
            } else if (parameters.file) {
                switch (parameters.file) {
                    case '.pipeline/default_pipeline_environment.yml':
                        return [default: 'config']
                    case '.pipeline/custom.yml':
                        return [custom: 'myConfig']
                    case 'pipeline_config.yml':
                        usedConfigFile = parameters.file
                        return [
                            general: [
                                productiveBranch: 'main'
                            ],
                            steps: [
                                mavenExecute: [
                                    dockerImage: 'my-custom-maven-docker']
                            ]
                        ]
                }
            } else {
                throw new IllegalArgumentException("Key 'text' and 'file' are both missing in map ${m}.")
            }
            usedConfigFile = parameters.file
            return yamlParser.load(examplePipelineConfig)
        })
        utilsMock = newUtilsMock()
    }

    Utils newUtilsMock() {
        def utilsMock = new Utils()
        utilsMock.steps = [
            stash  : { Map params -> pipelineAndTestStashIncludes = params.includes },
            unstash: {  }
        ]
        utilsMock.echo = { def m -> }
        return utilsMock
    }

    @Test
    void testIsYamlConfigurationAvailable() throws Exception {

        helper.registerAllowedMethod("fileExists", [String], { String path ->
            return path.endsWith('.pipeline/config.yml')
        })

        stepRule.step.setupCommonPipelineEnvironment(script: nullScript, utils: utilsMock)

        assertEquals('.pipeline/config.yml', usedConfigFile)
        assertNotNull(nullScript.commonPipelineEnvironment.configuration)
        assertEquals('develop', nullScript.commonPipelineEnvironment.configuration.general.productiveBranch)
        assertEquals('my-maven-docker', nullScript.commonPipelineEnvironment.configuration.steps.mavenExecute.dockerImage)
        assertEquals('.pipeline/**', pipelineAndTestStashIncludes)
    }

    @Test
    void testWorksAlsoWithYamlFileEnding() throws Exception {

        helper.registerAllowedMethod("fileExists", [String], { String path ->
            return path.endsWith('.pipeline/config.yaml')
        })

        stepRule.step.setupCommonPipelineEnvironment(script: nullScript, utils: utilsMock)

        assertEquals('.pipeline/config.yaml', usedConfigFile)
        assertNotNull(nullScript.commonPipelineEnvironment.configuration)
        assertEquals('develop', nullScript.commonPipelineEnvironment.configuration.general.productiveBranch)
        assertEquals('my-maven-docker', nullScript.commonPipelineEnvironment.configuration.steps.mavenExecute.dockerImage)
        assertEquals('.pipeline/**', pipelineAndTestStashIncludes)
    }

    @Test
    void testWorksAlsoWithCustomConfig() throws Exception {

        helper.registerAllowedMethod("fileExists", [String], { String path ->
            return path.endsWith('pipeline_config.yml')
        })

        stepRule.step.setupCommonPipelineEnvironment(script: nullScript, utils: utilsMock, configFile: 'pipeline_config.yml')

        assertEquals('pipeline_config.yml', usedConfigFile)
        assertNotNull(nullScript.commonPipelineEnvironment.configuration)
        assertEquals('main', nullScript.commonPipelineEnvironment.configuration.general.productiveBranch)
        assertEquals('my-custom-maven-docker', nullScript.commonPipelineEnvironment.configuration.steps.mavenExecute.dockerImage)
        assertEquals('.pipeline/**, pipeline_config.yml', pipelineAndTestStashIncludes)
    }

    @Test
    void testAttemptToLoadNonExistingConfigFile() {

        helper.registerAllowedMethod("fileExists", [String], { String path ->
            switch(path) {
                case 'default_pipeline_environment.yml': return false
                case 'custom.yml': return false
                case 'notFound.yml': return false
                case '': throw new RuntimeException('cannot call fileExists with empty path')
                default: return true
            }
        })

        helper.registerAllowedMethod("handlePipelineStepErrors", [Map,Closure], { Map map, Closure closure ->
            closure()
        })

        // Behavior documented here based on reality check
        thrown.expect(hudson.AbortException.class)
        thrown.expectMessage('No such library resource notFound could be found')

        stepRule.step.setupCommonPipelineEnvironment(
            script: nullScript,
            utils: utilsMock,
            customDefaults: 'notFound.yml'
        )
    }

    @Test
    void testInvalidEntriesInCustomDefaults() {

        helper.registerAllowedMethod("fileExists", [String], { String path ->
            switch(path) {
                case 'default_pipeline_environment.yml': return false
                case '': throw new RuntimeException('cannot call fileExists with empty path')
                default: return true
            }
        })

        helper.registerAllowedMethod("handlePipelineStepErrors", [Map,Closure], { Map map, Closure closure ->
            closure()
        })

        helper.registerAllowedMethod("readYaml", [Map], { Map parameters ->
            Yaml yamlParser = new Yaml()
            if (parameters.text) {
                return yamlParser.load(parameters.text)
            } else if (parameters.file) {
                if (parameters.file == '.pipeline/config-with-custom-defaults.yml') {
                    return [customDefaults: ['', true]]
                }
            }
            throw new IllegalArgumentException("Unexpected invocation of readYaml step")
        })

        stepRule.step.setupCommonPipelineEnvironment(
            script: nullScript,
            utils: utilsMock,
            configFile: '.pipeline/config-with-custom-defaults.yml'
        )

        assertEquals('WARNING: Ignoring invalid entry in custom defaults from files: \'\' \n' +
            'WARNING: Ignoring invalid entry in custom defaults from files: \'true\' \n', loggingRule.getLog())
    }

    @Test
    void testAttemptToLoadFileFromURL() {
        helper.registerAllowedMethod("fileExists", [String], {String path ->
            switch (path) {
                case 'default_pipeline_environment.yml': return false
                case '': throw new RuntimeException('cannot call fileExists with empty path')
                default: return true
            }
        })

        String customDefaultUrl = "https://url-to-my-config.com/my-config.yml"
        boolean urlRequested = false

        helper.registerAllowedMethod("httpRequest", [Map], {Map parameters ->
            switch (parameters.url) {
                case customDefaultUrl:
                    urlRequested = true
                    return [status: 200, content: "custom: 'myRemoteConfig'"]
                default:
                    throw new IllegalArgumentException('wrong URL requested')
            }
        })

        helper.registerAllowedMethod("readYaml", [Map], { Map parameters ->
            Yaml yamlParser = new Yaml()
            if (parameters.text) {
                return yamlParser.load(parameters.text)
            } else if (parameters.file) {
                if (parameters.file == '.pipeline/config-with-custom-defaults.yml') {
                    return [customDefaults: "${customDefaultUrl}"]
                }
                if (parameters.file == '.pipeline/custom_default_from_url_0.yml') {
                    return [custom: 'myRemoteConfig']
                }
            }
            throw new IllegalArgumentException("Unexpected invocation of readYaml step")
        })

        stepRule.step.setupCommonPipelineEnvironment(
            script: nullScript,
            utils: utilsMock,
            customDefaults: 'custom.yml',
            configFile: '.pipeline/config-with-custom-defaults.yml',
        )
        assertEquals("custom: 'myRemoteConfig'", writeFileRule.files['.pipeline/custom_default_from_url_0.yml'])
        assertEquals('myRemoteConfig', DefaultValueCache.instance.defaultValues['custom'])
    }


    @Test
    void inferBuildToolMaven() {
        helper.registerAllowedMethod('fileExists', [String.class], { s ->
            return s == "pom.xml"
        })
        setupCommonPipelineEnvironment.inferBuildTool(nullScript, [inferBuildTool: true])
        assertEquals('maven', nullScript.commonPipelineEnvironment.buildTool)
    }

    @Test
    void inferBuildToolMTA() {
        helper.registerAllowedMethod('fileExists', [String.class], { s ->
            return s == "mta.yaml"
        })
        setupCommonPipelineEnvironment.inferBuildTool(nullScript, [inferBuildTool: true])
        assertEquals('mta', nullScript.commonPipelineEnvironment.buildTool)
    }

    @Test
    void inferBuildToolNpm() {
        helper.registerAllowedMethod('fileExists', [String.class], { s ->
            return s == "package.json"
        })
        setupCommonPipelineEnvironment.inferBuildTool(nullScript, [inferBuildTool: true])
        assertEquals('npm', nullScript.commonPipelineEnvironment.buildTool)
    }

    @Test
    void inferBuildToolNone() {
        helper.registerAllowedMethod('fileExists', [String.class], { s ->
            return false
        })
        setupCommonPipelineEnvironment.inferBuildTool(nullScript, [inferBuildTool: true])
        assertNull(nullScript.commonPipelineEnvironment.buildTool)
    }

    @Test
    void "Set scmInfo parameter sets commit id"() {

        def GitUtils gitUtils = new GitUtils() {
            boolean isMergeCommit(){
                return false
            }
        }

        helper.registerAllowedMethod("fileExists", [String], { String path ->
            return path.endsWith('.pipeline/config.yml')
        })

        def dummyScmInfo = [GIT_COMMIT: 'dummy_git_commit_id', GIT_URL: 'https://github.com/testOrg/testRepo.git']

        stepRule.step.setupCommonPipelineEnvironment(script: nullScript, utils: utilsMock, scmInfo: dummyScmInfo, gitUtils: gitUtils)
        assertThat(nullScript.commonPipelineEnvironment.gitCommitId, is('dummy_git_commit_id'))
    }

    @Test
    void "Set scmInfo parameter sets git reference for branch"() {

        def GitUtils gitUtils = new GitUtils() {
            boolean isMergeCommit(){
                return false
            }
        }

        helper.registerAllowedMethod("fileExists", [String], { String path ->
            return path.endsWith('.pipeline/config.yml')
        })

        def dummyScmInfo = [GIT_COMMIT: 'dummy_git_commit_id', GIT_BRANCH: 'origin/testbranch']

        stepRule.step.setupCommonPipelineEnvironment(script: nullScript, utils: utilsMock, scmInfo: dummyScmInfo, gitUtils: gitUtils)
        assertThat(nullScript.commonPipelineEnvironment.gitRef, is('refs/heads/testbranch'))
    }

    @Test
    void "Set scmInfo parameter sets git reference for branch with slashes in name, origin"() {

        def GitUtils gitUtils = new GitUtils() {
            boolean isMergeCommit(){
                return false
            }
        }

        helper.registerAllowedMethod("fileExists", [String], { String path ->
            return path.endsWith('.pipeline/config.yml')
        })

        def dummyScmInfo = [GIT_COMMIT: 'dummy_git_commit_id', GIT_BRANCH: 'origin/testbranch/001']

        stepRule.step.setupCommonPipelineEnvironment(script: nullScript, utils: utilsMock, scmInfo: dummyScmInfo, gitUtils: gitUtils)
        assertThat(nullScript.commonPipelineEnvironment.gitRef, is('refs/heads/testbranch/001'))
    }

    @Test
    void "Set scmInfo parameter sets git reference for branch with slashes in name, not origin"() {

        def GitUtils gitUtils = new GitUtils() {
            boolean isMergeCommit(){
                return false
            }
        }

        helper.registerAllowedMethod("fileExists", [String], { String path ->
            return path.endsWith('.pipeline/config.yml')
        })

        def dummyScmInfo = [GIT_COMMIT: 'dummy_git_commit_id', GIT_BRANCH: 'testbranch/001']

        stepRule.step.setupCommonPipelineEnvironment(script: nullScript, utils: utilsMock, scmInfo: dummyScmInfo, gitUtils: gitUtils)
        assertThat(nullScript.commonPipelineEnvironment.gitRef, is('refs/heads/testbranch/001'))
    }

    @Test
        void "Set scmInfo parameter sets git reference for tag"() {

            def GitUtils gitUtils = new GitUtils() {
                boolean isMergeCommit(){
                    return false
                }
            }

            helper.registerAllowedMethod("fileExists", [String], { String path ->
                return path.endsWith('.pipeline/config.yml')
            })

            def dummyScmInfo = [GIT_COMMIT: 'dummy_git_commit_id', GIT_BRANCH: 'refs/tags/tag-1.0.0']

            stepRule.step.setupCommonPipelineEnvironment(script: nullScript, utils: utilsMock, scmInfo: dummyScmInfo, gitUtils: gitUtils)
            assertThat(nullScript.commonPipelineEnvironment.gitRef, is('refs/tags/tag-1.0.0'))
        }

    @Test
    void "sets gitReference and gitRemoteCommit for pull request, head strategy"() {

        def GitUtils gitUtils = new GitUtils() {
            boolean isMergeCommit(){
                return false
            }

            String getMergeCommitSha(){
                return "dummy_merge_git_commit_id"
            }
        }

        helper.registerAllowedMethod("fileExists", [String], { String path ->
            return path.endsWith('.pipeline/config.yml')
        })

        def dummyScmInfo = [GIT_COMMIT: 'dummy_git_commit_id', GIT_BRANCH: 'PR-42']

        stepRule.step.setupCommonPipelineEnvironment(script: nullScript, utils: utilsMock, scmInfo: dummyScmInfo, gitUtils: gitUtils)
        assertThat(nullScript.commonPipelineEnvironment.gitRef, is('refs/pull/42/head'))
        assertThat(nullScript.commonPipelineEnvironment.gitRemoteCommitId, is('dummy_git_commit_id'))
    }

    @Test
    void "sets gitReference and gitRemoteCommit for pull request, merge strategy"() {

        def GitUtils gitUtils = new GitUtils() {
            boolean isMergeCommit(){
                return true
            }

            String getMergeCommitSha(){
                return "dummy_merge_git_commit_id"
            }
        }

        helper.registerAllowedMethod("fileExists", [String], { String path ->
            return path.endsWith('.pipeline/config.yml')
        })

        def dummyScmInfo = [GIT_COMMIT: 'dummy_git_commit_id', GIT_BRANCH: 'PR-42']

        stepRule.step.setupCommonPipelineEnvironment(script: nullScript, utils: utilsMock, scmInfo: dummyScmInfo, gitUtils: gitUtils)
        assertThat(nullScript.commonPipelineEnvironment.gitRef, is('refs/pull/42/merge'))
        assertThat(nullScript.commonPipelineEnvironment.gitRemoteCommitId, is('dummy_merge_git_commit_id'))
    }

    @Test
    void "Set merge commit id as NA on exception"() {

        def GitUtils gitUtils = new GitUtils() {
            boolean isMergeCommit(){
                return true
            }

            String getMergeCommitSha() throws MissingPropertyException{
                throw new MissingPropertyException('pullRequest Context not found')
            }
        }

        helper.registerAllowedMethod("fileExists", [String], { String path ->
            return path.endsWith('.pipeline/config.yml')
        })

        def dummyScmInfo = [GIT_COMMIT: 'dummy_git_commit_id', GIT_BRANCH: 'PR-42']

        stepRule.step.setupCommonPipelineEnvironment(script: nullScript, utils: utilsMock, scmInfo: dummyScmInfo, gitUtils: gitUtils)
        assertThat(nullScript.commonPipelineEnvironment.gitRef, is('refs/pull/42/merge'))
        assertThat(nullScript.commonPipelineEnvironment.gitRemoteCommitId, is('NA'))
    }

    @Test
    void "No scmInfo passed as parameter yields empty git info"() {
        helper.registerAllowedMethod("fileExists", [String], { String path ->
            return path.endsWith('.pipeline/config.yml')
        })

        stepRule.step.setupCommonPipelineEnvironment(script: nullScript, utils: utilsMock)
        assertNull(nullScript.commonPipelineEnvironment.gitCommitId)
        assertNull(nullScript.commonPipelineEnvironment.getGitSshUrl())
        assertNull(nullScript.commonPipelineEnvironment.getGitHttpsUrl())
        assertNull(nullScript.commonPipelineEnvironment.getGithubOrg())
        assertNull(nullScript.commonPipelineEnvironment.getGithubRepo())
        assertNull(nullScript.commonPipelineEnvironment.getGitRef())
    }

    @Test
    void testSetScmInfoOnCommonPipelineEnvironment() {
        //currently supported formats
        def scmInfoTestList = [
            [GIT_URL: 'https://github.com/testOrg/testRepo.git', expectedSsh: 'git@github.com:testOrg/testRepo.git', expectedHttp: 'https://github.com/testOrg/testRepo.git', expectedOrg: 'testOrg', expectedRepo: 'testRepo'],
            [GIT_URL: 'https://github.com:7777/testOrg/testRepo.git', expectedSsh: 'git@github.com:testOrg/testRepo.git', expectedHttp: 'https://github.com:7777/testOrg/testRepo.git', expectedOrg: 'testOrg', expectedRepo: 'testRepo'],
            [GIT_URL: 'git@github.com:testOrg/testRepo.git', expectedSsh: 'git@github.com:testOrg/testRepo.git', expectedHttp: 'https://github.com/testOrg/testRepo.git', expectedOrg: 'testOrg', expectedRepo: 'testRepo'],
            [GIT_URL: 'ssh://git@github.com/testOrg/testRepo.git', expectedSsh: 'ssh://git@github.com/testOrg/testRepo.git', expectedHttp: 'https://github.com/testOrg/testRepo.git', expectedOrg: 'testOrg', expectedRepo: 'testRepo'],
            [GIT_URL: 'ssh://git@github.com:7777/testOrg/testRepo.git', expectedSsh: 'ssh://git@github.com:7777/testOrg/testRepo.git', expectedHttp: 'https://github.com/testOrg/testRepo.git', expectedOrg: 'testOrg', expectedRepo: 'testRepo'],
            [GIT_URL: 'ssh://git@github.com/path/to/testOrg/testRepo.git', expectedSsh: 'ssh://git@github.com/path/to/testOrg/testRepo.git', expectedHttp: 'https://github.com/path/to/testOrg/testRepo.git', expectedOrg: 'path/to/testOrg', expectedRepo: 'testRepo'],
            [GIT_URL: 'ssh://git@github.com/path/testOrg/testRepo.git', expectedSsh: 'ssh://git@github.com/path/testOrg/testRepo.git', expectedHttp: 'https://github.com/path/testOrg/testRepo.git', expectedOrg: 'path/testOrg', expectedRepo: 'testRepo'],
            [GIT_URL: 'ssh://git@github.com/testRepo.git', expectedSsh: 'ssh://git@github.com/testRepo.git', expectedHttp: 'https://github.com/testRepo.git', expectedOrg: 'N/A', expectedRepo: 'testRepo'],
        ]

        scmInfoTestList.each {scmInfoTest ->
            stepRule.step.setupCommonPipelineEnvironment.setGitUrlsOnCommonPipelineEnvironment(nullScript, scmInfoTest.GIT_URL)
            assertThat(nullScript.commonPipelineEnvironment.getGitSshUrl(), is(scmInfoTest.expectedSsh))
            assertThat(nullScript.commonPipelineEnvironment.getGitHttpsUrl(), is(scmInfoTest.expectedHttp))
            assertThat(nullScript.commonPipelineEnvironment.getGithubOrg(), is(scmInfoTest.expectedOrg))
            assertThat(nullScript.commonPipelineEnvironment.getGithubRepo(), is(scmInfoTest.expectedRepo))
        }
    }
}
