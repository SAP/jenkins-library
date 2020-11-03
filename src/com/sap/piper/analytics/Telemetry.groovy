package com.sap.piper.analytics

import com.cloudbees.groovy.cps.NonCPS

import org.jenkinsci.plugins.workflow.steps.FlowInterruptedException

class Telemetry implements Serializable{

    protected static Telemetry instance

    protected List listenerList = []

    protected Telemetry(){}

    protected static Telemetry getInstance(){
        if(!instance) {
            instance = new Telemetry()

            registerListener({ steps, payload ->
                piperOsDefaultReporting(steps, payload)
            })
        }
        return instance
    }

    static void registerListener(Closure listener){
        getInstance().listenerList.add(listener)
    }

    static notify(Script steps, Map config, Map payload){
        //allow opt-out via configuration
        if (!config?.collectTelemetryData) {
            steps.echo "[${payload.step}] Sending telemetry data is disabled."
            return
        }

        getInstance().listenerList.each { listener ->
            try {
                listener(steps, payload)
            } catch (ignore) {
                // some error occurred in telemetry reporting. This should not break anything though.
                steps.echo "[${payload.step}] Telemetry Report with listener failed: ${ignore.getMessage()}"
            }
        }
    }

    protected static void piperOsDefaultReporting(Script steps, Map payload) {
        def swaEndpoint = 'https://webanalytics.cfapps.eu10.hana.ondemand.com/tracker/log'
        Map swaPayload = [
            'idsite': '827e8025-1e21-ae84-c3a3-3f62b70b0130',
            'url': 'https://github.com/SAP/jenkins-library',
            'action_name': payload.actionName,
            'event_type': payload.eventType,
            'custom3': payload.step,            // custom3 = step name (passed as parameter step)
            'custom4': payload.jobUrlSha1,      // custom4 = job url hashed (calculated)
            'custom5': payload.buildUrlSha1,    // custom5 = build url hashed (calculated)
            'custom10': payload.stageName       // custom10 = stage name
        ]
        // step related parameters
        for(def key : [1, 2, 3, 4, 5]){         // custom11 - custom15 = step related parameter 1 - 5 (passed as parameter stepParam1 - stepParam5)
            if (payload["stepParam${key}"] != null) swaPayload.put("custom1${key}", payload["stepParam${key}"])
        }

        try {
            steps.timeout(
                time: 10,
                unit: 'SECONDS'
            ){
                steps.httpRequest(url: "${swaEndpoint}?${getPayloadString(swaPayload)}", timeout: 5, quiet: true)
            }
        } catch (FlowInterruptedException ignore){
            // telemetry reporting timed out. This should not break anything though.
            steps.echo "[${payload.step}] Telemetry Report with listener failed: timeout"
        }
    }

    @NonCPS
    private static String getPayloadString(Map payload){
        return payload
            .collect { entry -> return "${entry.key}=${URLEncoder.encode(entry.value.toString(), "UTF-8")}" }
            .join('&')
    }
}
