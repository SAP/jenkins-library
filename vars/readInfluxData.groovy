import static com.sap.piper.Prerequisites.checkScript
import com.sap.piper.analytics.InfluxData

import groovy.transform.Field

@Field def STEP_NAME = getClass().getName()

void call(Map parameters = [:]) {
    final script = checkScript(this, parameters) ?: this
    String piperGoPath = parameters?.piperGoPath ?: './piper'
    def output = script.sh(returnStdout: true, script: "${piperGoPath} readInfluxData${parameters.verbose?' --verbose':''}")

    InfluxData.readFromJson(script, output)
}
