pipeline {
    agent { docker 'golang:latest' }
    stages {
        stage('Build') {
            steps {
                sh 'rm -f nac.syso'
                sh 'apk add libpcap-dev && go build'
            }
        }
    }
}