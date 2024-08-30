import com.sap.piper.ConfigurationHelper
import com.sap.piper.DownloadCacheUtils
import com.sap.piper.GenerateDocumentation
import com.sap.piper.k8s.ContainerMap
import groovy.transform.Field
import com.sap.piper.Utils

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/npmExecuteEndToEndTests.yaml'

void call(Map parameters = [:]) {
    List credentials = []
    final script = checkScript(this, parameters) ?: this
    piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials)
}

