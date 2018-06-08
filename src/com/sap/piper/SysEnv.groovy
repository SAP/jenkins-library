package com.sap.piper

class SysEnv implements Serializable {
    static final long serialVersionUID = 1L

    private Map env

    List<String> envNames=[
        'HTTP_PROXY',
        'HTTPS_PROXY',
        'NO_PROXY',
        'http_proxy',
        'https_proxy',
        'no_proxy'
    ]

    public SysEnv() {
        env= new HashMap<String,String>()
        fillMap()
    }

    public String get(String key) {
        return env.get(key)
    }

    public Map getEnv() {
        return env
    }

    public String remove(String key) {
        return env.remove(key)
    }

    @NonCPS
    private void fillMap() {
        for (String name in envNames) {
            if(System.getenv(name)){
                env.put(name,System.getenv(name))
            }
        }
    }
}
