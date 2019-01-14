import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.*

import static org.hamcrest.Matchers.*
import static org.junit.Assert.assertThat

class ContainerExecuteStructureTestsTest extends BasePiperTest {
    private ExpectedException thrown = ExpectedException.none()
    private JenkinsStepRule jsr = new JenkinsStepRule(this)
    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)
    private JenkinsShellCallRule jscr = new JenkinsShellCallRule(this)
    private JenkinsDockerExecuteRule jedr = new JenkinsDockerExecuteRule(this)

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(thrown)
        .around(jedr)
        .around(jscr)
        .around(jlr)
        .around(jsr) // needs to be activated after jedr, otherwise executeDocker is not mocked

    @Before
    void init() throws Exception {
        helper.registerAllowedMethod('stash', [String.class], null)
        helper.registerAllowedMethod("findFiles", [Map.class], { map ->
            def files
            if(map.glob == 'notFound.json')
                files = []
            else if(map.glob == 'cst/*.yml')
                files = [
                    new File("cst/test1.yml"),
                    new File("cst/test2.yml")
                ]
            else
                files = [new File(map.glob)]
            return files.toArray()
        })
    }

    @Test
    void testExecuteContainterStructureTestsDefault() throws Exception {
        helper.registerAllowedMethod('readFile', [String.class], {s ->
            return '{testResult: true}'
        })
        jsr.step.containerExecuteStructureTests(
            script: nullScript,
            juStabUtils: utils,
            containerCommand: '/busybox/tail -f /dev/null',
            containerShell: '/busybox/sh',
            dockerImage: 'myRegistry:55555/pathTo/myImage:myTag',
            testConfiguration: 'cst/*.yml',
            testImage: 'myRegistry/myImage:myTag'
        )
        // asserts
        assertThat(jscr.shell, hasItem(allOf(
            stringContainsInOrder(['#!/busybox/sh', 'container-structure-test', '--config']),
            containsString('--config cst\\test1.yml'),
            containsString('--config cst\\test2.yml'),
            containsString('--driver docker'),
            containsString('--image myRegistry/myImage:myTag'),
            containsString('--test-report ./cst-report.json'),
        )))
        //currently no default Docker image
        assertThat(jedr.dockerParams.dockerImage, is('myRegistry:55555/pathTo/myImage:myTag'))
        assertThat(jedr.dockerParams.dockerOptions, is("-u 0 --entrypoint=''"))
        assertThat(jedr.dockerParams.containerCommand, is('/busybox/tail -f /dev/null'))
        assertThat(jedr.dockerParams.containerShell, is('/busybox/sh'))
        assertThat(jlr.log, containsString('{testResult: true}'))
        assertThat(jscr.shell, hasItem('docker pull myRegistry/myImage:myTag'))
    }

    @Test
    void testExecuteContainterStructureTestsK8S() throws Exception {
        def envDefault = nullScript.env
        nullScript.env = [ON_K8S: 'true']
        jsr.step.containerExecuteStructureTests(
            script: nullScript,
            juStabUtils: utils,
            dockerImage: 'myRegistry:55555/pathTo/myImage:myTag',
            testConfiguration: 'cst/*.yml',
            testImage: 'myRegistry/myImage:myTag'
        )
        nullScript.env = envDefault
        // asserts
        assertThat(jscr.shell, hasItem(allOf(
            stringContainsInOrder(['#!/bin/sh', 'container-structure-test', '--config']),
            containsString('--config cst\\test1.yml'),
            containsString('--config cst\\test2.yml'),
            containsString('--driver tar'),
            containsString('--image myRegistry/myImage:myTag'),
            containsString('--test-report ./cst-report.json'),
        )))
        assertThat(jedr.dockerParams.dockerImage, is('myRegistry:55555/pathTo/myImage:myTag'))
        assertThat(jscr.shell, not(hasItem('docker pull myRegistry/myImage:myTag')))
    }

    @Test
    void testExecuteContainterStructureTestsError() throws Exception {
        helper.registerAllowedMethod('readFile', [String.class], {s ->
            return '{testResult: true}'
        })
        helper.registerAllowedMethod('sh', [String.class], {s ->
            if (s.startsWith('#!/busybox/sh\ncontainer-structure-test test')) {
                throw new GroovyRuntimeException('shell call failed')
            } else {
                return null
            }
        })
        thrown.expectMessage('shell call failed')

        jsr.step.containerExecuteStructureTests(
            script: nullScript,
            juStabUtils: utils,
            containerCommand: '/busybox/tail -f /dev/null',
            containerShell: '/busybox/sh',
            dockerImage: 'myRegistry:55555/pathTo/myImage:myTag',
            testConfiguration: 'cst/*.yml',
            testImage: 'myRegistry/myImage:myTag'
        )
    }

    @Test
    void testExecuteContainterStructureTestsErrorNoFailure() throws Exception {
        helper.registerAllowedMethod('readFile', [String.class], {s ->
            return '{testResult: true}'
        })
        helper.registerAllowedMethod('sh', [String.class], {s ->
            if (s.startsWith('#!/busybox/sh\ncontainer-structure-test test')) {
                throw new GroovyRuntimeException('shell call failed')
            } else {
                return null
            }
        })

        jsr.step.containerExecuteStructureTests(
            script: nullScript,
            juStabUtils: utils,
            containerCommand: '/busybox/tail -f /dev/null',
            containerShell: '/busybox/sh',
            dockerImage: 'myRegistry:55555/pathTo/myImage:myTag',
            failOnError: false,
            testConfiguration: 'cst/*.yml',
            testImage: 'myRegistry/myImage:myTag'
        )

        assertThat(jlr.log, containsString('Test execution failed'))
    }
}
