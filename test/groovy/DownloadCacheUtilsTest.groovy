import com.sap.piper.DownloadCacheUtils
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.JenkinsEnvironmentRule
import util.JenkinsFileExistsRule
import util.JenkinsShellCallRule
import util.Rules

import static org.junit.Assert.assertEquals
import static org.junit.Assert.assertThat
import static org.junit.Assert.assertTrue

class DownloadCacheUtilsTest extends BasePiperTest{
    private JenkinsShellCallRule shellRule = new JenkinsShellCallRule(this)
    private JenkinsFileExistsRule fileExistsRule = new JenkinsFileExistsRule(this, [])

    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(shellRule)
        .around(fileExistsRule)

    @Before
    void init() {
        helper.registerAllowedMethod("libraryResource", [String.class], { path ->
            File resource = new File(new File('resources'), path)
            if (resource.exists()) {
                return resource.getText()
            }
            return ''
        })

    }


    @Test
    void writeGlobalMavenSettingsForDownloadCacheShouldWriteFile() {
        //binding.variables.env.DL_CACHE_HOSTNAME = 'cx-downloadcache'
        binding.setVariable('env', [DL_CACHE_HOSTNAME: 'cx-downloadcache'])
        helper.registerAllowedMethod('node', [String.class, Closure.class]) {s, body ->
            body()
        }

        String expected = '.pipeline/global_settings.xml'
        String actual = DownloadCacheUtils.writeGlobalMavenSettingsForDownloadCache(nullScript)

        assertEquals(expected, actual)
    }

    @Test
    void writeGlobalMavenSettingsForDownloadCacheShouldNotWriteFile() {
        fileExistsRule.registerExistingFile('.pipeline/global_settings.xml')
    }

    @Test
    void writeGlobalMavenSettingsForDownloadCacheShouldReturnEmptyStringOnNoDlCache() {

    }
}
