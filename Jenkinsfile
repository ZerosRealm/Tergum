pipeline {
  agent {
    docker {
      image 'golang:latest'
    }

  }

  environment {
    XDG_CACHE_HOME = '/tmp/.cache'
    PROJECT_NAME="Tergum"
    DOMAIN="zerosrealm.xyz"
    STACK="tergum"
    DOCKER_REGISTRY="https://registry.zerosrealm.xyz"
    CONTAINER="zerosrealm/tergum"
    VERSION="1.${BUILD_NUMBER}"
  }

  stages {
    stage('Build') {
      steps {
        sh 'go build cmd/server/server.go'
        sh 'go build cmd/agent/agent.go'
        echo 'Building done.'
      }
    }

    stage('Build Server Image') {
      steps {
        script {
            docker.withRegistry("${DOCKER_REGISTRY}", "zerosregistry-creds") {
                def img = docker.build("${CONTAINER}:${VERSION}", "./dockerfiles/server")
                img.push('latest')
                sh "docker rmi ${img.id}"
            }
        }
      }
    }

    stage('Build Agent Image') {
      steps {
        script {
            docker.withRegistry("${DOCKER_REGISTRY}", "zerosregistry-creds") {
                def img = docker.build("${CONTAINER}-agent:${VERSION}", "./dockerfiles/agent")
                img.push('latest')
                sh "docker rmi ${img.id}"
            }
        }
      }
    }

    stage('Deploy') {
      steps {
        withCredentials([string(credentialsId: 'tergum-deploy', variable: 'DEPLOY_URL')]) {
          script {
            echo "Deploying Container Stack"
            sh 'curl -X POST $DEPLOY_URL'
          }
          echo 'Deployed!'
        }
      }
    }

  }
}
