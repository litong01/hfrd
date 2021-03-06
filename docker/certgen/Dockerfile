FROM alpine/git AS BUILD

RUN cd /tmp && git init hfrd && cd hfrd                            && \
    git remote add origin https://github.com/litong01/hfrd.git     && \
    git config core.sparsecheckout true                            && \
    echo "docker/certgen/*" >> .git/info/sparse-checkout        && \
    git pull --depth=1 origin master

FROM ubuntu:18.04

MAINTAINER Tong Li <email4tong@gmail.com>

ARG user=ubuntu
ARG uid=1000
ARG gid=1000

RUN apt-get update -y                                                && \
    apt-get install -y bash xxd python-pip sudo                        && \
    pip install --upgrade pip ansible pyyaml                       && \
    groupadd -g ${gid} ${user}                                     && \
    useradd -d /opt/agent -u ${uid} -g ${user} ${user}             && \
    usermod -a -G root ${user}                                     && \
    echo "${user} ALL=(ALL) NOPASSWD: ALL"|tee /etc/sudoers.d/${user} && \
    mkdir -p /opt/agent/.ssh                                       && \
    cd /opt/agent/.ssh                                             && \
    echo "host *" > config                                         && \
    echo "  StrictHostKeyChecking no" >> config                    && \
    echo "  UserKnownHostsFile /dev/null" >> config

COPY --from=build /tmp/hfrd/docker/certgen /opt/agent
ADD https://storage.googleapis.com/kubernetes-release/release/v1.15.0/bin/linux/amd64/kubectl /usr/local/bin/kubectl
RUN chown -R ${uid}:${gid} /opt/agent                              && \
    chmod 755 /usr/local/bin/kubectl

ENV HOME /opt/agent
ENV WORKDIR /opt/agent
WORKDIR /opt/agent
USER ${user}

CMD [ "ansible-playbook", "--version" ]