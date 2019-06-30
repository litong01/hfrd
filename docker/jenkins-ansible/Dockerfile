FROM jenkins/jenkins:lts
USER root
ENV JAVA_OPTS="-Djenkins.install.runSetupWizard=false"
ADD https://storage.googleapis.com/kubernetes-release/release/v1.11.2/bin/linux/amd64/kubectl /usr/local/bin/kubectl
ADD security.groovy /usr/share/jenkins/ref/init.groovy.d/security.groovy
ADD plugins.txt /usr/share/jenkins/ref/plugins.txt
ADD jenkins.ini /usr/share/hfrd/jenkins.ini
ADD hfrd-jenkins-jobs.yaml /usr/share/hfrd/jjb/hfrd-jenkins-jobs.yaml
RUN apt-get update && apt-get install -y bash                                  && \
    apt-get install -y python-pip openssh-client sudo apt-transport-https         \
    ca-certificates curl software-properties-common                            && \
    pip install --upgrade pip ansible pyyaml jenkins-job-builder               && \
    chmod +x /usr/local/bin/kubectl                                            && \
    chown jenkins:jenkins /usr/share/jenkins/ref/init.groovy.d/security.groovy && \
    chown jenkins:jenkins /usr/share/jenkins/ref/plugins.txt                   && \
    /usr/local/bin/install-plugins.sh < /usr/share/jenkins/ref/plugins.txt     && \
    curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add -    && \
    add-apt-repository                                                            \
    "deb [arch=amd64] https://download.docker.com/linux/ubuntu bionic stable"  && \
    apt-get update -qq && apt-get install -qqy docker-ce                       && \
    usermod -a -G docker jenkins                                               && \
    echo 'jenkins ALL=(ALL) NOPASSWD: ALL' >f> /etc/sudoers.d/jenkins           && \
    rm -rf /var/lib/apt/lists/*

VOLUME /var/lib/docker
USER jenkins
