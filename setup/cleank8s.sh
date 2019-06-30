#!/bin/sh

# Remove all daemon set
all=$(kubectl --kubeconfig kubeconfig get daemonset -o name)
if [ ! -z "$all" ]; then
  echo "Removing daemon sets..."
  kubectl --kubeconfig kubeconfig delete $all
else
  echo "No daemon sets to remove"
fi

# Remove all the services
all=$(kubectl --kubeconfig kubeconfig get services -o name|grep -v "service/kubernetes")
if [ ! -z "$all" ]; then
  echo "Removing all the services ..."
  kubectl --kubeconfig kubeconfig delete $all
else
  echo "No services to remove"
fi

# Remove all stateful sets
all=$(kubectl --kubeconfig kubeconfig get statefulsets -o name)
if [ ! -z "$all" ]; then
  echo "Removing statefulsets..."
  kubectl --kubeconfig kubeconfig delete $all
else
  echo "No statefulsets to remove"
fi

# Remove all pods
all=$(kubectl --kubeconfig kubeconfig get pods -o name)
if [ ! -z "$all" ]; then
  echo "Removing fabric network pods..."
  kubectl --kubeconfig kubeconfig delete $all
else
  echo "No pods to remove"
fi

all=$(kubectl --kubeconfig kubeconfig get jobs -o name)
if [ ! -z "$all" ]; then
  echo "Removing jobs..."
  kubectl --kubeconfig kubeconfig delete $all
else
  echo "No jobs to remove"
fi

# Remove all the PVCs that jobs may have created
# echo "Removing PVCs..."
# all=$(kubectl --kubeconfig kubeconfig get pvc -o name)
# kubectl --kubeconfig kubeconfig delete $all
