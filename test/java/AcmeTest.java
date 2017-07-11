import org.apache.commons.io.FileUtils;
import org.apache.commons.lang.StringUtils;
import org.jenkinsci.plugins.workflow.cps.CpsFlowDefinition;
import org.jenkinsci.plugins.workflow.cps.global.UserDefinedGlobalVariable;
import org.jenkinsci.plugins.workflow.job.WorkflowJob;
import org.jenkinsci.plugins.workflow.job.WorkflowRun;
import org.junit.Test;
import org.junit.runners.model.Statement;

import java.io.File;
import java.util.Arrays;


public class AcmeTest extends AbstractJenkinsTest {

    /**
     * Test acme getter and setter
     */
    @Test
    public void acmeTest() throws Exception {
        story.addStep(new Statement() {
            @Override public void evaluate() throws Throwable {
                //File vars = new File(repo.workspace, UserDefinedGlobalVariable.PREFIX);
                File vars = new File(repo.workspace, "vars");
                vars.mkdirs();
                FileUtils.writeStringToFile(new File(vars, "acme.groovy"), StringUtils.join(Arrays.asList(
                        "class acme implements Serializable {",
                        "   private String name = 'initial'",
                        "   def setName(value) {",
                        "       this.name = value",
                        "   }",
                        "   def getName() {",
                        "       this.name",
                        "   }",
                        "   def caution(message) {",
                        "       echo \"Hello, ${name}! CAUTION: ${message}\"",
                        "   }",
                        "}")
                        , "\n"));

                // simulate the effect of push
                uvl.rebuild();

                WorkflowJob p = jenkins.createProject(WorkflowJob.class, "p");

                p.setDefinition(new CpsFlowDefinition(
                        "node {\n" +

                                "acme.setName('acmeName')\n"+
                                "echo acme.getName()\n" +

                                "}",
                        true));

                // build this workflow
                WorkflowRun b = story.j.assertBuildStatusSuccess(p.scheduleBuild2(0));

                story.j.assertLogContains("acmeName", b);
            }
        });
    }

    //@Test
    public void acmeTest2() throws Exception {
        story.addStep(new Statement() {
            @Override
            public void evaluate() throws Throwable {

                copyLibrarySources();
                // simulate the effect of push
                uvl.rebuild();

                WorkflowJob p = jenkins.createProject(WorkflowJob.class, "p");

                p.setDefinition(new CpsFlowDefinition(
                        "import com.sap.piper.Utils\n" +
                                "node {\n" +

                                "acme.setName('myName')\n"+
                                "assert acme.getName() == 'myName'\n" +


                                "}",
                        true));

                // build this workflow
                WorkflowRun b = story.j.assertBuildStatusSuccess(p.scheduleBuild2(0));
            }
        });
    }
 }
