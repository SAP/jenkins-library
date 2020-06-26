import com.sap.piper.ConfigurationLoader

def call(Map parameters) {
    def script = parameters.script
    def feature = parameters.feature
    ConfigurationLoader.generalConfiguration(script).features?.get(feature) ?: false
}
