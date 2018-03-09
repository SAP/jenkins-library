package com.sap.piper

import hudson.AbortException
import java.io.File


class FileUtils implements Serializable {

    static directoryOrFileExists(script, dirOrFile) {
        if (!dirOrFile) throw new IllegalArgumentException("The parameter 'dirOrFile' can not be null or empty.")
        def returnStatus = script.sh returnStatus: true, script: """
                                                                 #!/bin/bash --login

                                                                 if [ -d $dirOrFile ]; then
                                                                   echo \"$dirOrFile exists.\"
                                                                   exit 0
                                                                 elif [ -f $dirOrFile ]; then
                                                                   echo \"$dirOrFile exists.\"
                                                                   exit 0
                                                                 else
                                                                   echo \"$dirOrFile does not exist.\"
                                                                   exit 1
                                                                 fi
                                                                 """
        if (returnStatus == 0) return true
        else return false
    }

    static isDirectory(script, dir) {
        if (!dir) throw new IllegalArgumentException("The parameter 'dir' can not be null or empty.")
        def returnStatus = script.sh returnStatus: true, script: """
                                                                 #!/bin/bash --login

                                                                 if [ -d $dir ]; then
                                                                   echo \"$dir is a directory.\"
                                                                   exit 0
                                                                 else
                                                                   echo \"$dir is not a directory.\"
                                                                   exit 0
                                                                 fi
                                                                 """
        if (returnStatus == 0) return true
        else return false
    }

    static isDirectoryEmpty(script, dir) {
        if (!dir) throw new IllegalArgumentException("The parameter 'dir' can not be null or empty.")
        def returnStatus = script.sh returnStatus: true, script: """
                                                               #!/bin/bash --login

                                                               if [ -z "\$(ls -A $dir)" ]; then
                                                                 echo "$dir is empty."
                                                                 exit 1
                                                               else
                                                                 echo "$dir is not empty."
                                                                 exit 0
                                                               fi
                                                               """
        if (returnStatus == 0) return false
        else return true
    }

    static isFile(script, filePath) {
        if (!filePath) throw new IllegalArgumentException("The parameter 'filePath' can not be null or empty.")
        validateDirectoryOrFileExists(script, filePath)
        def returnStatus = script.sh returnStatus: true, script: """
                                                               #!/bin/bash --login

                                                               if [ -f $filePath ]; then
                                                                 echo \"$filePath is a file.\"
                                                                 exit 0
                                                               else
                                                                 echo \"$filePath is not a file.\"
                                                                 exit 1
                                                               fi
                                                               """
        if (returnStatus == 0) return true
        else return false
    }

    static validateDirectoryOrFileExists(script, dirOrFile) {
        if (!dirOrFile) throw new IllegalArgumentException("The parameter 'dirOrFile' can not be null or empty.")
        if (!directoryOrFileExists(script, dirOrFile)) throw new AbortException("Validation failed. '$dirOrFile' does not exist.")
    }

    static validateDirectory(script, dir) {
        if (!dir) throw new IllegalArgumentException("The parameter 'dir' can not be null or empty.")
        validateDirectoryOrFileExists(script, dir)
        if (!isDirectory(script, dir)) throw new AbortException("Validation failed. '$dir' is not a directory.")
    }

    static validateDirectoryIsNotEmpty(script, dir) {
        validateDirectory(script, dir)
        if (isDirectoryEmpty(script, dir)) throw new AbortException("Validation failed. '$dir' is empty.")
    }

    static validateFile(script, filePath) {
        if (!filePath) throw new IllegalArgumentException("The parameter 'filePath' can not be null or empty.")
        validateDirectoryOrFileExists(script, filePath)
        if (!isFile(script, filePath)) throw new AbortException("Validation failed. '$filePath' is not a file.")
    }
}
