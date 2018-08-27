package com.sap.piper.cm

import java.util.Map

import com.sap.piper.GitUtils

import hudson.AbortException


public class ChangeManagement implements Serializable {

    private script
    private GitUtils gitUtils

    public ChangeManagement(def script, GitUtils gitUtils = null) {
        this.script = script
        this.gitUtils = gitUtils ?: new GitUtils()
    }

    String getChangeDocumentId(
                              String from = 'origin/master',
                              String to = 'HEAD',
                              String label = 'ChangeDocument\\s?:',
                              String format = '%b'
                            ) {

        return getLabeledItem('ChangeDocumentId', from, to, label, format)
    }

    String getTransportRequestId(
                              String from = 'origin/master',
                              String to = 'HEAD',
                              String label = 'TransportRequest\\s?:',
                              String format = '%b'
                            ) {

        return getLabeledItem('TransportRequestId', from, to, label, format)
    }

    private String getLabeledItem(
                              String name,
                              String from,
                              String to,
                              String label,
                              String format
                            ) {

        if( ! gitUtils.insideWorkTree() ) {
            throw new ChangeManagementException("Cannot retrieve ${name}. Not in a git work tree. ${name} is extracted from git commit messages.")
        }

        def items = gitUtils.extractLogLines(".*${label}.*", from, to, format)
                                .collect { line -> line?.replaceAll(label,'')?.trim() }
                                .unique()

        items.retainAll { line -> line != null && ! line.isEmpty() }

        if( items.size() == 0 ) {
            throw new ChangeManagementException("Cannot retrieve ${name} from git commits. ${name} retrieved from git commit messages via pattern '${label}'.")
        } else if (items.size() > 1) {
            throw new ChangeManagementException("Multiple ${name}s found: ${items}. ${name} retrieved from git commit messages via pattern '${label}'.")
        }

        return items[0]
    }

    boolean isChangeInDevelopment(String changeId, String endpoint, String credentialsId, String clientOpts = '') {
        int rc = executeWithCredentials(endpoint, credentialsId, 'is-change-in-development', [new KeyValue('cID', changeId), new Value('--return-code').setQuotes(false)],
            clientOpts) as int

        if (rc == 0) {
            return true
        } else if (rc == 3) {
            return false
        } else {
            throw new ChangeManagementException("Cannot retrieve status for change document '${changeId}'. Does this change exist? Return code from cmclient: ${rc}.")
        }
    }

    String createTransportRequest(String changeId, String developmentSystemId, String endpoint, String credentialsId, String clientOpts = '') {
        try {
            def transportRequest = executeWithCredentials(endpoint, credentialsId, 'create-transport', [new KeyValue('cID', changeId), new KeyValue('dID', developmentSystemId)],
                clientOpts)
            return transportRequest.trim() as String
        }catch(AbortException e) {
            throw new ChangeManagementException("Cannot create a transport request for change id '$changeId'. $e.message.")
        }
    }


    void uploadFileToTransportRequest(String changeId, String transportRequestId, String applicationId, String filePath, String endpoint, String credentialsId, String cmclientOpts = '') {
        int rc = executeWithCredentials(endpoint, credentialsId, 'upload-file-to-transport', [new KeyValue('cID', changeId),
                                                                                                 new KeyValue('tID', transportRequestId),
                                                                                                 new Value(applicationId).setQuotes(false), new Value(filePath)],
            cmclientOpts) as int

        if(rc == 0) {
            return
        } else {
            throw new ChangeManagementException("Cannot upload file '$filePath' for change document '$changeId' with transport request '$transportRequestId'. Return code from cmclient: $rc.")
        }

    }
    
    def executeWithCredentials(String endpoint, String credentialsId, String command, List<CLOption> args, String clientOpts = '') {
        script.withCredentials([script.usernamePassword(
            credentialsId: credentialsId,
            passwordVariable: 'password',
            usernameVariable: 'username')]) {
            def returnValue = script.sh(returnStatus: true,
                script: getCMCommandLine(endpoint, script.username, script.password,
                    command, args,
                    clientOpts))
            return returnValue;

        }
    }

    void releaseTransportRequest(String changeId, String transportRequestId, String endpoint, String credentialsId, String clientOpts = '') {
        int rc = executeWithCredentials( endpoint, credentialsId, 'release-transport', [new KeyValue('cID', changeId),
                                                                                        new KeyValue('tID', transportRequestId)], clientOpts) as int
        if(rc == 0) {
            return
        } else {
            throw new ChangeManagementException("Cannot release Transport Request '$transportRequestId'. Return code from cmclient: $rc.")
        }
    }
    
    String getCMCommandLine(String endpoint,
                            String username,
                            String password,
                            String command,
                            List<CLOption> args,
                            String clientOpts = '') {
        String cmCommandLine = '#!/bin/bash'
        if(clientOpts) {
            cmCommandLine +=  """
                             export CMCLIENT_OPTS="${clientOpts}" """
        }
        
        cmCommandLine += """
                        cmclient -e '$endpoint' \
                           -u '$username' \
                           -p '$password' \
                           -t SOLMAN \
                          ${command} ${args.collect{it}.join(' ')}
                    """
        
        return cmCommandLine
    }
    
    abstract static class CLOption 
    {
        String key
        String value
        boolean quoteValue=true;
        
        protected CLOption(String key, String value) {
            this.key = key
            this.value = value
        }
        CLOption setQuotes(boolean quoteValue) {
            this.quoteValue=quoteValue
            return this
        }
        String toString() {
            StringBuilder  sb=new StringBuilder()
            if(key!=null) {
                sb.append("-${key}")
            }
            if(value!=null) {
                if(sb.length()!=0) {
                    sb.append(" ")
                }
                if(quoteValue) {
                    sb.append("'$value'")
                }
                else {
                    sb.append(value)
                }
            }
            return sb.toString()
        }
    }
    
    public static final class KeyValue extends CLOption     {
        KeyValue(String key, String value) {
            super(key,value)
            if(key==null) throw new NullPointerException("Option name is missing")
            if(value==null) throw new NullPointerException("Option value is missing")
        }
    }
    public static final class Switch extends CLOption {
        Switch(String key) {
            super(key,null)
            if(key==null) throw new NullPointerException("Option name is missing")
        }
    }
    public static final class Value extends CLOption{
        Value(String value) {
            super(null,value)
            if(value==null) throw new NullPointerException("Option value is missing")
        }
    }
}
