import com.lesfurets.jenkins.unit.BasePipelineTest

import static ProjectSource.projectSource
import static com.lesfurets.jenkins.unit.global.lib.LibraryConfiguration.library

import org.junit.Rule
import org.junit.rules.TemporaryFolder

public class PiperTestBase extends BasePipelineTest {

    @Rule
    public TemporaryFolder pipelineFolder = new TemporaryFolder()

    private File pipeline

    protected messages = [], shellCalls = []

    void setUp() {

        super.setUp()

        messages.clear()
        shellCalls.clear()

        preparePiperLib()

        helper.registerAllowedMethod('echo', [String], {s -> messages.add(s)} )
        helper.registerAllowedMethod('sh', [String], {  s ->
            shellCalls.add(s.replaceAll(/\s+/, " ").trim())
        })

        pipeline = pipelineFolder.newFile()
    }

    protected withPipeline(p) {
        pipeline << p()
        loadScript(pipeline.toURI().getPath())
    }

    private preparePiperLib() {
        def piperLib = library()
            .name('piper-library-os')
            .retriever(projectSource())
            .targetPath('clonePath/is/not/necessary')
            .defaultVersion('<irrelevant>')
            .allowOverride(true)
            .implicit(false)
            .build()
        helper.registerSharedLibrary(piperLib)
    }
}
