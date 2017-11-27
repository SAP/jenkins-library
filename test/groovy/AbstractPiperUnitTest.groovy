import com.lesfurets.jenkins.unit.BasePipelineTest
import com.sap.piper.DefaultValueCache
import org.junit.Before
import org.yaml.snakeyaml.Yaml

import static com.lesfurets.jenkins.unit.global.lib.LibraryConfiguration.library

class AbstractPiperUnitTest extends BasePipelineTest {

    @Before
    void setUp() {
        super.setUp()

        DefaultValueCache.reset()

        def library = library()
            .name('piper-library-os')
            .retriever(ProjectSource.projectSource())
            .targetPath('clonePath/is/not/necessary')
            .defaultVersion("master")
            .allowOverride(true)
            .implicit(false)
            .build()
        helper.registerSharedLibrary(library)

        helper.registerAllowedMethod("readYaml", [Map], { Map parameters ->
            Yaml yamlParser = new Yaml()
            return yamlParser.load(parameters.text)
        })
    }
}
