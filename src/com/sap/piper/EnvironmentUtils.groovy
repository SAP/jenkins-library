package com.sap.piper

import hudson.AbortException


class EnvironmentUtils implements Serializable {

    def static isEnvironmentVariable(script, variable) {
        def envVar
        try {
          envVar = script.sh returnStdout: true, script: """#!/bin/bash --login
                                                            echo \$$variable"""
        } catch(AbortException e) {
          throw new AbortException("The verification of the environment variable '$variable' failed. Reason: $e.message.")
        }
        if (envVar.trim()) return true
        else return false
    }

    def static getEnvironmentVariable(script, variable) {
        try {
          def envVar = script.sh returnStdout: true, script: """#!/bin/bash --login
                                                                echo \$$variable"""
          return envVar.trim()
        } catch(AbortException e) {
          throw new AbortException("There was an error requesting the environment variable '$variable'. Reason: $e.message.")
        }
    }
}
