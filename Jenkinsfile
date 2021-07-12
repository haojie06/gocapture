pipeline {
    agent { docker 'golang:latest' }
    stages {
        stage('build') {
            steps {
                sh 'apk add libpcap-dev'
                sh 'rm -f nac.syso'
                sh 'go build'
            }
        }
    }
}