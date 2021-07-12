pipeline {
    agent { docker 'golang:latest' }
    stages {
        stage('build') {
            steps {
                sh 'rm -f nac.syso'
                sh 'go build'
            }
        }
    }
}