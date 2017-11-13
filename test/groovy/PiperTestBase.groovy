import com.lesfurets.jenkins.unit.BasePipelineTest

import static ProjectSource.projectSource
import static com.lesfurets.jenkins.unit.global.lib.LibraryConfiguration.library

public class PiperTestBase extends BasePipelineTest {

    protected final void _setUp() {
        super.setUp()
        preparePiperLib()
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
