# Additional
## Transport
          stage('prepare') {
              checkout scm
              setupCommonPipelineEnvironment script:this
              checkChangeInDevelopment script: this
          }


