import org.junit.Rule;
import org.junit.Test
import org.junit.rules.RuleChain;

import com.sap.piper.DefaultValueCache

import util.BasePiperTest
import util.JenkinsStepRule;
import util.Rules

public class PrepareDefaultValuesTest extends BasePiperTest {

    private JenkinsStepRule jsr = new JenkinsStepRule(this)

    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(jsr)

	@Test
	public void testMerge() {

		helper.registerAllowedMethod("libraryResource", [String], { fileName-> return fileName })
		helper.registerAllowedMethod("readYaml", [Map], { m ->
			switch(m.text) {
				case 'default_pipeline_environment.yml': return [a: 'x']
				default: return [the:'end']
			}
		})
		jsr.step.call(script: nullScript)
		
		
		println DefaultValueCache.getInstance().getDefaultValues()
	}
}
