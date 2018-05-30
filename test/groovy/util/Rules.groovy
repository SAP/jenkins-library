package util

import org.junit.rules.RuleChain

import com.lesfurets.jenkins.unit.BasePipelineTest
import com.lesfurets.jenkins.unit.global.lib.LibraryConfiguration

public class Rules {

    public static RuleChain getCommonRules(BasePipelineTest testCase) {
        return getCommonRules(testCase, null)
    }

    public static RuleChain getCommonRules(BasePipelineTest testCase, LibraryConfiguration libConfig) {
        return RuleChain.outerRule(new JenkinsSetupRule(testCase, libConfig))
		    .around(new JenkinsReadYamlRule(testCase))
            .around(new JenkinsResetDefaultCacheRule())
            .around(new JenkinsErrorRule(testCase))
			.around(new JenkinsEnvironmentRule(testCase))
    }
}
