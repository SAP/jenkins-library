import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.*

import static org.hamcrest.Matchers.containsString
import static org.junit.Assert.*

class PipelineStashFilesBeforeBuildTest extends BasePiperTest {
    JenkinsStepRule jsr = new JenkinsStepRule(this)
    JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)
    JenkinsShellCallRule jscr = new JenkinsShellCallRule(this)
    //JenkinsReadJsonRule jrj = new JenkinsReadJsonRule(this)

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        //.around(jrj)
        .around(jlr)
        .around(jscr)
        .around(jsr)

    @Test
    void testStashBeforeBuildNoOpa() {

        jsr.step.call(script: nullScript, juStabUtils: utils)

        // asserts
        assertEquals('mkdir -p gitmetadata', jscr.shell[0])
        assertEquals('cp -rf .git/* gitmetadata', jscr.shell[1])
        assertEquals('chmod -R u+w gitmetadata', jscr.shell[2])
        assertFalse(jlr.log.contains('Stash content: opa5'))
        assertThat(jlr.log, containsString('Stash content: git (include: **/gitmetadata/**, exclude: )'))
        assertThat(jlr.log, containsString('Stash content: tests (include: **/pom.xml, **/*.json, **/*.xml, **/src/**, **/node_modules/**, **/specs/**, **/env/**, **/*.js, exclude: )'))
        assertThat(jlr.log, containsString('Stash content: buildDescriptor (include: **/pom.xml, **/.mvn/**, **/assembly.xml, **/.swagger-codegen-ignore, **/package.json, **/requirements.txt, **/setup.py, **/whitesource_config.py, **/mta*.y*ml, **/.npmrc, **/whitesource.*.json, **/whitesource-fs-agent.config, .xmake.cfg, Dockerfile, **/VERSION, **/version.txt, **/build.sbt, **/sbtDescriptor.json, **/project/*, exclude: **/node_modules/**/package.json)'))
        assertThat(jlr.log, containsString('Stash content: deployDescriptor (include: **/manifest*.y*ml, **/*.mtaext.y*ml, **/*.mtaext, **/xs-app.json, helm/**, *.y*ml, exclude: )'))
        assertThat(jlr.log, containsString('Stash content: opensource configuration (include: **/srcclr.yml'))
        assertThat(jlr.log, containsString('Stash content: snyk configuration (include: **/.snyk'))
        assertThat(jlr.log, containsString('Stash content: pipelineConfigAndTests (include: .pipeline/*.*'))
        assertThat(jlr.log, containsString('Stash content: securityDescriptor (include: **/xs-security.json'))
    }

    @Test
    void testStashBeforeBuildOpa() {

        jsr.step.call(script: nullScript, juStabUtils: utils, runOpaTests: true)

        // asserts
        assertThat(jlr.log, containsString('Stash content: opa5'))
        assertThat(jlr.log, containsString('Stash content: git'))
        assertThat(jlr.log, containsString('Stash content: tests'))
        assertThat(jlr.log, containsString('Stash content: buildDescriptor'))
        assertThat(jlr.log, containsString('Stash content: deployDescriptor'))
        assertThat(jlr.log, containsString('Stash content: opensource configuration'))
        assertThat(jlr.log, containsString('Stash content: snyk configuration'))
        assertThat(jlr.log, containsString('Stash content: pipelineConfigAndTests'))
        assertThat(jlr.log, containsString('Stash content: securityDescriptor'))
    }

    @Test
    void testStashBeforeBuildOpaCompatibility() {

        jsr.step.call(script: nullScript, juStabUtils: utils, runOpaTests: 'true')

        // asserts
        assertThat(jlr.log, containsString('Stash content: opa5'))
        assertThat(jlr.log, containsString('Stash content: git'))
        assertThat(jlr.log, containsString('Stash content: tests'))
        assertThat(jlr.log, containsString('Stash content: buildDescriptor'))
        assertThat(jlr.log, containsString('Stash content: deployDescriptor'))
        assertThat(jlr.log, containsString('Stash content: opensource configuration'))
        assertThat(jlr.log, containsString('Stash content: snyk configuration'))
        assertThat(jlr.log, containsString('Stash content: pipelineConfigAndTests'))
        assertThat(jlr.log, containsString('Stash content: securityDescriptor'))
    }
}
