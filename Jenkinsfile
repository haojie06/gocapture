pipeline {
    agent { docker 'golang:latest' }
    stages {
        stage('Build') {
            steps {
                sh 'apt update'
                sh 'rm -f nac.syso'
                sh 'apt install -y libpcap-dev && go build'
            }
        }
    }
}