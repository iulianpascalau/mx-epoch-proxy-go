pipeline {
    agent any

    parameters {
        string(name: 'BRANCH_OR_TAG', defaultValue: 'main', description: 'Branch or Tag to deploy (e.g., main, v1.0.0)')
    }

    environment {
        PROD_VM_USER = 'ubuntu'
        PROD_VM_HOST = credentials('epoch-proxy-prod-vm-ip')
        PROJECT_PATH = '/home/ubuntu/epoch-proxy'
    }

    stages {
        stage('Deploy') {
            steps {
                sshagent(['prod-vm-ssh-key']) {
                     sh "ssh -o StrictHostKeyChecking=no ${PROD_VM_USER}@${PROD_VM_HOST} 'cd ${PROJECT_PATH} && ./scripts/deploy.sh ${params.BRANCH_OR_TAG}'"
                }
            }
        }
    }
}
