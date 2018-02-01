package util

import com.lesfurets.jenkins.unit.BasePipelineTest
import org.junit.rules.TestRule
import org.junit.runner.Description
import org.junit.runners.model.Statement

class JenkinsReadMavenPomRule implements TestRule {

    final BasePipelineTest testInstance
    final String testRoot

    List shell = []

    JenkinsReadMavenPomRule(BasePipelineTest testInstance, String testRoot) {
        this.testInstance = testInstance
        this.testRoot = testRoot
    }

    @Override
    Statement apply(Statement base, Description description) {
        return statement(base)
    }

    private Statement statement(final Statement base) {
        return new Statement() {
            @Override
            void evaluate() throws Throwable {

                testInstance.helper.registerAllowedMethod('readMavenPom', [Map.class], {m -> return loadPom("${testRoot}/${m.file}")})

                base.evaluate()
            }
        }
    }

    MockPom loadPom( String path ){
        return new MockPom( path )
    }

    class MockPom {
        def f
        def pom
        MockPom(String path){
            this.f = new File( path )
            if ( f.exists() ){
                this.pom = new XmlSlurper().parse(f)
            }
            else {
                throw new FileNotFoundException( 'Failed to find file: ' + path )
            }
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
