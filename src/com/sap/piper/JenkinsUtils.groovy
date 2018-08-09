package com.sap.piper

import com.cloudbees.groovy.cps.NonCPS

@NonCPS
def isPluginActive(pluginId) {
    return Jenkins.instance.pluginManager.plugins.find { p -> p.isActive() && p.getShortName() == pluginId }
}
