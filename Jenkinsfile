pipeline {
    agent { docker 'golang:latest' }
    stages {
        stage('build') {
            steps {
                sh 'go build'
            }
        }
    }
}