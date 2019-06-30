#!/usr/bin/env groovy
timestamps {
    node ('hfrd') {
        properties([disableConcurrentBuilds()])
        stage('checkout scm') {
            checkout scm
        }

        commitHash = sh(returnStdout: true, script: 'git rev-parse HEAD').trim().take(7)
        commitText = sh(returnStdout: true, script: 'git show -s --format=format:"*%s*  _by %an_" HEAD').trim()
        changeLog = "`${commitHash}` ${commitText}"
        try {
            stage('hfrd/server docker build test') {
                sh '''set -e
                      make api-docker
                   '''
            }

            stage('test modules test against fabric v1.2') {
                sh '''set -e
                      cd modules/gosdk
                      ./gosdk_example.sh 1.2
                   '''
            }
            stage('test modules test against fabric v1.3') {
                sh '''set -e
                      cd modules/gosdk
                      ./gosdk_example.sh 1.3
                   '''
            }
            stage('test modules test against fabric v1.4') {
                sh '''set -e
                      cd modules/gosdk
                      ./gosdk_example.sh 1.4
                   '''
            }
        } catch(e) {
            slackSend color: 'danger', message: "@here FAILED: Job '${env.JOB_NAME} [${env.BUILD_NUMBER}]' (${env.BUILD_URL})\n${changeLog}"
            throw e
        }
        slackSend color: 'good', message: "SUCCESSFUL:  Job '${env.JOB_NAME} [${env.BUILD_NUMBER}]' (${env.BUILD_URL})\n${changeLog}"
    }
}
