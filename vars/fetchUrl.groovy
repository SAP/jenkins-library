import com.cloudbees.groovy.cps.NonCPS

@NonCPS
def call(url) {
    return new URL(url).getText()
}