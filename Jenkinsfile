pipeline {
  agent {
    docker {
      image 'golang:latest'
    }

  }
  stages {
    stage('Build') {
      steps {
        sh 'go build cmd/server/server.go'
        sh 'go build cmd/agent/agent.go'
        echo 'Building done.'
      }
    }

    stage('Deploy') {
      steps {
        echo 'Yay!\\nDeploying it!'
      }
    }

  }
}