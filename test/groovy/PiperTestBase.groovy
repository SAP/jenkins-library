import com.lesfurets.jenkins.unit.BasePipelineTest

import org.yaml.snakeyaml.Yaml

import static ProjectSource.projectSource
import static com.lesfurets.jenkins.unit.global.lib.LibraryConfiguration.library

import org.junit.Rule
import org.junit.rules.TemporaryFolder

public class PiperTestBase extends BasePipelineTest {

    @Rule
    public TemporaryFolder pipelineFolder = new TemporaryFolder()

    private File pipeline

    void setUp() {

        super.setUp()

        pipeline = pipelineFolder.newFile()

    }

    protected withPipeline(p) {
        pipeline << p
        loadScript(pipeline.toURI().getPath())
    }
}
