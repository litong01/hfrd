#!/bin/bash
NAME=${NAME:-${1}}
set -x
if [ -z ${NAME} ]; then
	echo -e "\nError: Release name is not set !!!\n"
	echo -e "Usage:\n\tNAME=<release-name> ./cleanup.sh ( OR )\n\t./cleanup <release-name>\n"
	exit
fi
rm -rf ${NAME}
kubectl  delete services $(kubectl get services | grep ${NAME}- | awk '{print $1}')
kubectl  delete pods  $(kubectl get pods | grep ${NAME}- | awk '{print $1}')
helm list --tls --all ${NAME}- | grep ${NAME}- | awk '{print $1}' | xargs helm delete --purge --tls
kubectl get secrets | grep ${NAME}- | awk '{print $1}' | xargs kubectl delete secret
kubectl get pvc | grep ${NAME}- | awk '{print $1}' | xargs kubectl delete pvc
set +x

echo "Waiting for PVC's to be removed"

SECONDS=0

while (( $SECONDS < 60 ));
do
    sleep 1
    kubectl get pvc | grep ${NAME}-
    RC=$(echo $?)
    if [ "$RC" == "1" ]; then
        echo "PVC's have been removed"
        break
    fi
    echo "Waiting for PVC's to be removed"
done

if [ $SECONDS -ge 60 ]
then
    echo "Timed out waiting for PVC's to be removed"
    exit 1
fi
