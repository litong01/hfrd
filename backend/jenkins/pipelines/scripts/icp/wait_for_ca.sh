#!/bin/bash

echo "Waiting for ${NAME} CA to come up"

SECONDS=0

while (( $SECONDS < 600 ));
do
    POD_STATUS=$(kubectl get pods | grep "${NAME}" | awk '{print $3}')
    IS_READY=$(kubectl get pods | grep "${NAME}" | awk '{print $2}')
    TOTAL_PODS=$(echo $IS_READY | cut -d'/' -f2)

        if [ "${IS_READY}" == "${TOTAL_PODS}/${TOTAL_PODS}" ]; then
            curl -s -k --connect-timeout 1 https://${CA_HOST}:${CA_PORT}/cainfo
            RET=$?
            if [ "$RET" == "0" ]; then
                echo "CA ${NAME} is running"
                break
            fi
        fi
        echo "Waiting for pod ${NAME} to start completion. Status = ${POD_STATUS}, Readiness = ${IS_READY}"
        sleep 3
done

if [ $SECONDS -ge 600 ]
then
    echo "Timed out waiting for ${NAME} CA to come up"
    exit 1
fi

