pipeline {
  triggers {
    pollSCM('') // Enabling being build on Push
  }
  environment {
    XDG_CACHE_HOME = '/tmp/.cache'
    PROJECT_NAME = 'Tergum'
    DOMAIN = 'zerosrealm.xyz'
    STACK = 'tergum'
    DOCKER_REGISTRY = 'https://registry.zerosrealm.xyz'
    CONTAINER = 'registry.zerosrealm.xyz/zerosrealm/tergum'
    VERSION = "1.${BUILD_NUMBER}"
  }
  agent {
    label 'master'
  }
  stages {
    stage('Build binaries') {
      agent {
        docker {
          image 'golang:latest'
        }

      }
      steps {
        sh 'go build cmd/server/server.go'
        sh 'go build cmd/agent/agent.go'
        echo 'Building done.'
      }
    }

    stage('Build Server Image') {
      agent {
        docker {
          image 'docker:latest'
        }
      }
      options { skipDefaultCheckout() }
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

    stage('Build Agent Image') {
      agent {
        docker {
          image 'docker:latest'
        }
      }
      options { skipDefaultCheckout() }
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
      agent any
      steps {
        withCredentials(bindings: [string(credentialsId: 'tergum-deploy', variable: 'DEPLOY_URL')]) {
          script {
            echo "Deploying Container Stack"
            sh 'curl -X POST $DEPLOY_URL'
          }

          echo 'Deployed!'
        }

      }
    }
  }
  post {
      always {
          cleanWs(
                  cleanWhenNotBuilt: false,
                  deleteDirs: true,
                  disableDeferredWipeout: true,
                  notFailBuild: true,
                  patterns: [[pattern: '.gitignore', type: 'INCLUDE'],
                  [pattern: '.propsfile', type: 'EXCLUDE']]
          )
      }
  }
}
