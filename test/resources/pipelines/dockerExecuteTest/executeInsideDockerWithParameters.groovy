@Library('piper-library-os')

execute() {
    node() {
        dockerExecute(script: this, dockerImage: 'maven:3.5-jdk-8-alpine', dockerOptions: '-it', dockerVolumeBind: ['my_vol': '/my_vol'], dockerEnvVars: ['http_proxy': 'http://proxy:8000']) {
            echo 'Inside Docker'
        }
    }
}

return this
