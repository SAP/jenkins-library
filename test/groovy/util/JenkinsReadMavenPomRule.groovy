package util

import com.lesfurets.jenkins.unit.BasePipelineTest
import org.junit.rules.TestRule
import org.junit.runner.Description
import org.junit.runners.model.Statement
import java.nio.file.Path
import java.nio.file.Paths

class JenkinsReadMavenPomRule implements TestRule {

    final BasePipelineTest testInstance
    final String testRoot
    def poms = [:]

    JenkinsReadMavenPomRule(BasePipelineTest testInstance, String testRoot) {
        this.testInstance = testInstance
        this.testRoot = testRoot
    }

    JenkinsReadMavenPomRule registerPom(fileName, pom) {
        poms.put(fileName, pom)
        return this
    }

        @Override
    Statement apply(Statement base, Description description) {
        return statement(base)
    }

    private Statement statement(final Statement base) {
        return new Statement() {
            @Override
            void evaluate() throws Throwable {

                testInstance.helper.registerAllowedMethod('readMavenPom', [Map.class], { Map m ->
                    if(m.text) {
                        return loadPomfromText(m.text)
                    } else if(m.file) {
                        def pom = poms.get(m.file)
                        if(pom == null)
                            return loadPom("${testRoot}/${m.file}")
                        if(pom instanceof Closure) pom = pom()
                            return loadPomfromText(pom)
                    } else {
                        throw new IllegalArgumentException("Key 'text' and 'file' are both missing in map ${m}.")
                    }
                })

                base.evaluate()
            }
        }
    }

    MockPom loadPom( String path ){
        return new MockPom(Paths.get(path))
    }

    MockPom loadPomfromText( String text){
        return new MockPom( text )
    }

    class MockPom {
        def pom
        MockPom(Path path){
            def f = new File( path )
            if ( f.exists() ){
                this.pom = new XmlSlurper().parse(f)
            }
            else {
                throw new FileNotFoundException( 'Failed to find file: ' + path )
            }
        }
        MockPom(String text ){
            this.pom = new XmlSlurper().parseText(text)
        }
        String getVersion(){
            return pom.version
        }
        String getGroupId(){
            return pom.groupId
        }
        String getArtifactId(){
            return pom.artifactId
        }
        String getPackaging(){
            return pom.packaging
        }
        String getName(){
            return pom.name
        }
    }
}
