package com.sap.piper.analytics

import com.cloudbees.groovy.cps.NonCPS
import org.jenkinsci.plugins.workflow.steps.MissingContextVariableException

class Analytics implements Serializable{

    private static Analytics instance

    private List eventListeners

    private Analytics(){
        this.eventListeners = []
    }

    @NonCPS
    static Analytics getInstance(){
        if(!instance) {
            createInstance()

            Closure defaultListener = {payload ->
                piperOsDefaultReporting(payload)
            }
            registerEventListener(defaultListener)
        }
        return instance
    }

    static void createInstance(){
        instance = new Analytics()
    }

    static notify(Script steps, Map config, Map payload){
        //allow opt-out via configuration
        if (!config?.collectTelemetryData) {
            steps.echo "[${payload.get('step')}] telemetry reporting disabled!"
            return
        }

        getInstance().eventListeners.each { listener ->
            try {
                listener(steps, payload)
            } catch (err) {
                // some error occured in telemetry reporting. This should not break anything though.
            }
        }
    }

    void registerEventListener(Closure listener){
        eventListeners.add(listener)
    }

    void piperOsDefaultReporting(Script steps, Map payload) {
        try {

            def swaCustom = [:]

            /* SWA custom parameters:
                custom3 = step name (passed as parameter step)
                custom4 = job url hashed (calculated)
                custom5 = build url hashed (calculated)
                custom10 = stage name
                custom11 = step related parameter 1 (passed as parameter stepParam1)
                custom12 = step related parameter 2 (passed as parameter stepParam2)
                custom13 = step related parameter 3 (passed as parameter stepParam3)
                custom14 = step related parameter 4 (passed as parameter stepParam4)
                custom15 = step related parameter 5 (passed as parameter stepParam5)
            */

            def swaUrl = 'https://webanalytics.cfapps.eu10.hana.ondemand.com/tracker/log'
            def idsite = '827e8025-1e21-ae84-c3a3-3f62b70b0130'
            def url = 'https://github.com/SAP/jenkins-library'

            def action_name = payload.actionName
            def event_type = payload.eventType

            swaCustom.custom3 = payload.step
            swaCustom.custom4 =  payload.jobUrlSha1
            swaCustom.custom5 = payload.buildUrlSha1
            swaCustom.custom10 = payload.stageName
            swaCustom.custom11 = payload.stepParam1
            swaCustom.custom12 = payload.stepParam2
            swaCustom.custom13 = payload.stepParam3
            swaCustom.custom14 = payload.stepParam4
            swaCustom.custom15 = payload.stepParam5

            def options = []
            options.push("-G")
            options.push("-v \"${swaUrl}\"")
            options.push("--data-urlencode \"action_name=${action_name}\"")
            options.push("--data-urlencode \"idsite=${idsite}\"")
            options.push("--data-urlencode \"url=${url}\"")
            options.push("--data-urlencode \"event_type=${event_type}\"")
            for(def key : ['custom3', 'custom4', 'custom5', 'custom10', 'custom11', 'custom12', 'custom13', 'custom14', 'custom15']){
                if (swaCustom[key] != null) options.push("--data-urlencode \"${key}=${swaCustom[key]}\"")
            }
            options.push("--connect-timeout 5")
            options.push("--max-time 20")

            steps.sh(returnStatus: true, script: "#!/bin/sh +x\ncurl ${options.join(' ')} > /dev/null 2>&1 || echo '[${payload.step}] Telemetry Report to SWA failed!'")

        } catch (MissingContextVariableException noNode) {
            echo "[${payload.step}] Telemetry Report to SWA skipped, no node available!"
        }
    }
}
