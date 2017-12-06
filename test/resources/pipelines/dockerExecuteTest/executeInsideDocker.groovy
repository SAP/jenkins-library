@Library('piper-library-os')

execute() {
    node() {
        dockerExecute(script: this, dockerImage: 'maven:3.5-jdk-8-alpine') {
            echo 'Inside Docker'
        }
    }
}

return this
