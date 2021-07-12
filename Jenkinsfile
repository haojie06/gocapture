pipeline {
    agent { docker 'golang:latest' }
    stages {
        stage('Build') {
            steps {
                sh 'rm -f nac.syso'
                sh 'apt update && apt install -y libpcap-dev && go build'
            }
        }
    }
}