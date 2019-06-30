# How to access k8s dashboard from localhost

This doc only applies to kubernetes clusters that are set up by our own.

IBM Cloud Kubernetes Service does NOT apply to this doc.

## Prerequisites
- kubernetes dashboard installed
- kubeconfig
- kubectl

## Steps
1. Put kubeconfig file in the right directory

Move kubeconfig file to ${HOME}/.kube/config

Refer to [kubernetes doc](https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/)

The below section is an sample of ${HOME}/.kube/config file content
```
apiVersion: v1
clusters:
- cluster:
    insecure-skip-tls-verify: true
    server: https://localhost:6443
  name: docker-for-desktop-cluster
- cluster:
    certificate-authority-data: xxxx
    server: https://9.37.200.207:6443
  name: rtp
contexts:
- context:
    cluster: docker-for-desktop-cluster
    user: docker-for-desktop
  name: docker-for-desktop
- context:
    cluster: rtp
    user: rtp
  name: rtp
current-context: rtp
kind: Config
preferences: {}
users:
- name: docker-for-desktop
  user:
    client-certificate-data: xxxx
    client-key-data: xxxx
- name: rtp
  user:
    client-certificate-data: xxxx
    client-key-data: xxxx
```

2. Start a new terminal to start kubectl proxy
```shell
xixuejia-mbp:~ xixuejia$ kubectl proxy
Starting to serve on 127.0.0.1:8001
```
3. Open browser to access k8s dashboard
You should be able to access k8s dashboard with this [link](http://localhost:8001/api/v1/namespaces/kube-system/services/https:kubernetes-dashboard:/proxy/#!/overview?namespace=default)