import java.io.File;
import org.apache.commons.io.FileUtils;
import org.jenkinsci.plugins.workflow.cps.CpsFlowDefinition;
import org.jenkinsci.plugins.workflow.job.WorkflowJob;
import org.jenkinsci.plugins.workflow.job.WorkflowRun;
import org.junit.runners.model.Statement;

public class _TemplateTest extends AbstractJenkinsTest {
    /**
     * Test ... step
     */
    //@Test
    public void testWhatEver() throws Exception {
        story.addStep(new Statement() {
            @Override
            public void evaluate() throws Throwable {
                copyLibrarySources();
                // simulate the effect of push
                uvl.rebuild();
                WorkflowJob p = jenkins.createProject(WorkflowJob.class, "p");
                //copy test resources into workspace
                FileUtils.copyDirectory(new File("test/resources"), new File(jenkins.getWorkspaceFor(p).getRemote(), "resources"));

                p.setDefinition(new CpsFlowDefinition(
                        "node {\n" +

                                "\n" +
                                "\n" +

                        "}",
                        true)
                );

                // build this workflow
                WorkflowRun b = story.j.assertBuildStatusSuccess(p.scheduleBuild2(0));
                //story.j.assertLogContains("this is part of the log", b);
            }
        });
    }
 }