import groovy.transform.Field
import static com.sap.piper.Prerequisites.checkScript
import com.sap.piper.DownloadCacheUtils

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/mtaBuild.yaml'

void call(Map parameters = [:]) {
    final script = checkScript(this, parameters) ?: this
    parameters = DownloadCacheUtils.injectDownloadCacheInMavenParameters(script, parameters)

    piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, [])
}
