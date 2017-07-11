import java.io.File;
import java.io.IOException;
import javax.inject.Inject;
import jenkins.model.Jenkins;
import org.apache.commons.io.FileUtils;
import org.jenkinsci.plugins.workflow.cps.global.UserDefinedGlobalVariableList;
import org.jenkinsci.plugins.workflow.cps.global.WorkflowLibRepository;
import org.junit.ClassRule;
import org.junit.Rule;
import org.jvnet.hudson.test.BuildWatcher;
import org.jvnet.hudson.test.RestartableJenkinsRule;

public class AbstractJenkinsTest {
	@ClassRule
	public static BuildWatcher buildWatcher = new BuildWatcher();
	@Rule
	public RestartableJenkinsRule story = new RestartableJenkinsRule();
	@Inject
	protected Jenkins jenkins;
	@Inject
	WorkflowLibRepository repo;
	@Inject
	protected UserDefinedGlobalVariableList uvl;

	public AbstractJenkinsTest() {
		super();
	}

	protected void copyLibrarySources() {
		try {
			FileUtils.copyDirectory(new File("vars"), new File(repo.workspace, "vars"));
			FileUtils.copyDirectory(new File("src"), new File(repo.workspace, "src"));
			FileUtils.copyDirectory(new File("resources"), new File(repo.workspace, "resources"));
		} catch (IOException e) {
			e.printStackTrace();
			System.exit(1);
		}
	}
}
