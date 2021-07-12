pipeline {
    agent { docker 'golang:latest' }
    stages {
        stage('Build') {
            steps {
                sh 'apt update && rm -f nac.syso && apt install -y libpcap-dev && go build'
            }
        }
    }
}