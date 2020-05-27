package util

import com.lesfurets.jenkins.unit.BasePipelineTest
import org.junit.rules.TestRule
import org.junit.runner.Description
import org.junit.runners.model.Statement

class JenkinsHttpRequestRule implements TestRule {

    final BasePipelineTest testInstance

    JenkinsHttpRequestRule(BasePipelineTest testInstance) {
        this.testInstance = testInstance
    }

    Map closures = [:]
    List requests = []

    void mockUrl(String url, Closure body){
        closures[url] = body
    }

    @Override
    Statement apply(Statement base, Description description) {
        return statement(base)
    }

    private Statement statement(final Statement base) {
        return new Statement() {
            @Override
            void evaluate() throws Throwable {

                testInstance.helper.helper.registerAllowedMethod('httpRequest', [Map], { p ->
                    requests.add(p)
                    if(p.url && closures.containsKey(p.url).toString()){
                        closures[p.url]()
                        return
                    }
                    throw new RuntimeException("Not implemented")
                })

                base.evaluate()
            }
        }
    }
}
