pipeline {
  agent {
      label 'docker1'
  }

  environment {
    XDG_CACHE_HOME = '/tmp/.cache'
    PROJECT_NAME="Tergum"
    DOMAIN="zerosrealm.xyz"
    STACK="tergum"
    DOCKER_REGISTRY="https://registry.zerosrealm.xyz"
    CONTAINER="registry.zerosrealm.xyz/zerosrealm/tergum"
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

    stage('Push Server Image') {
      steps {
        script {
            sh 'docker logout'
            withDockerRegistry(credentialsId: 'zerosregistry-creds', url: 'https://registry.zerosrealm.xyz/') {
                def img = docker.build("${CONTAINER}:${VERSION}", "-f ./dockerfiles/server .")
                img.push('latest')
                sh "docker rmi ${img.id}"
            }
        }
      }
    }

    stage('Push Agent Image') {
      steps {
        script {
            sh 'docker logout'
            withDockerRegistry(credentialsId: 'zerosregistry-creds', url: 'https://registry.zerosrealm.xyz/') {
                def img = docker.build("${CONTAINER}-agent:${VERSION}", "-f ./dockerfiles/agent .")
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
