#!/bin/sh

if [ ! -d "./kubeconfig" ]; then 
  echo -e "\nDirectory ./kubeconfig DOES NOT exist."
  if [ ! -z $KUBECONFIG ]; then
    echo "The environment variable KUBECONFIG is set to: ${KUBECONFIG}."
  else
    echo -e "The environment variable KUBECONFIG is not set.\n"
    exit 
  fi
else
  export KUBECONFIG=./kubeconfig
fi

## This MAY not pickup the right node if there are multiples in the pool
externalIP=$(kubectl describe nodes|grep "  ExternalIP:"|head -n 1|cut -d ':' -f2)
externalIP=$(echo $externalIP)

helm install ibm-hfrd --name hfrd --wait --values override.yaml --set nginx.allowOrigins="{http://localhost:8080,https://${externalIP}:30443}"

echo "Wait few seconds for pods to be created"
sleep 15
while : ; do
    res=$(kubectl logs hfrdjenkins | grep "Jenkins is fully up and running")
    if [[ ! -z $res ]]; then
        kubectl exec -it hfrdjenkins -- jenkins-jobs --conf /usr/share/hfrd/jenkins.ini update /usr/share/hfrd/jjb
        break
    fi
    echo "Waiting for jenkins to start up..."
    sleep 5
done

echo "Access service at https://${externalIP}:30443/hfrd/"