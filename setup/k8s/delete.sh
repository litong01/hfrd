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

helm delete --purge hfrd