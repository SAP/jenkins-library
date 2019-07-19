package com.sap.piper

class PathUtils implements Serializable {
    static final long serialVersionUID = 1L

    static String convertToAbsolutePath(def script, String path){
        if(path && !isAbsolutePath(path)){
            return script.env.WORKSPACE + "/" + path
        }
        return path
    }

    static isAbsolutePath(String path){
        return path && path.charAt(0) == '/'
    }
}
