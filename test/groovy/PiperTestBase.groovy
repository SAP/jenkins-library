import com.lesfurets.jenkins.unit.BasePipelineTest

import static ProjectSource.projectSource
import static com.lesfurets.jenkins.unit.global.lib.LibraryConfiguration.library

public class PiperTestBase extends BasePipelineTest {

    protected messages = [], shellCalls = []

    protected final void _setUp() {

        super.setUp()

        messages.clear()
        shellCalls.clear()

        preparePiperLib()

        helper.registerAllowedMethod('echo', [String], {s -> messages.add(s)} )
        helper.registerAllowedMethod('sh', [String], {  s ->
            shellCalls.add(s.replaceAll(/\s+/, " ").trim())
        })

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
