def source = 'https://binaries.sonarsource.com/Distribution/sonar-scanner-cli/sonar-scanner-cli-3.3.0.1492-linux.zip'
                def filename = source.tokenize('/').last()
                println filename.replace('.zip', '')