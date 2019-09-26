import com.sap.piper.ConfigurationLoader
import com.sap.piper.ConfigurationMerger
import com.sap.piper.CommonPipelineEnvironment
import com.sap.piper.analytics.InfluxData

class commonPipelineEnvironment implements Serializable {

    // We forward everything to the singleton instance of
    // commonPipelineEnvironment (CPE) on default value cache.
    //
    // Some background: each step has its own instance of CPE step.
    // In case each instance has its own set of properties these instances
    // are configured individually. Properties set on one instance cannot be
    // retrieved with another instance. Now each instance forwards to one singleton.
    // This means: all instances of the CPE shares the same properties/configuration.

    def methodMissing(String name, def args) {
        CommonPipelineEnvironment.getInstance().invokeMethod(name, args)
    }

    def propertyMissing(def name) {
        CommonPipelineEnvironment.getInstance()[name]
    }

    def propertyMissing(def name, def value) {
        CommonPipelineEnvironment.getInstance()[name] = value
    }
}
