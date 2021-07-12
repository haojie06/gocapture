pipeline {
    agent { docker 'golang:latest' }
    stages {
        stage('Build') {
            steps {
                sh 'rm -f nac.syso'
                sh 'apt install libpcap-dev && go build'
            }
        }
    }
}