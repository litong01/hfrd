#!/bin/bash
SECONDS=0

while (( $SECONDS < 600 ));
do
    POD_STATUS=$(kubectl get pods | grep "${NAME}" | awk '{print $3}')
    IS_READY=$(kubectl get pods | grep "${NAME}" | awk '{print $2}')
    TOTAL_PODS=$(echo $IS_READY | cut -d'/' -f2)
        echo "Waiting for pod ${NAME} to start completion. Status = ${POD_STATUS}, Readiness = ${IS_READY}"
        if [ "${IS_READY}" == "${TOTAL_PODS}/${TOTAL_PODS}" ]; then
            break
        fi
        sleep 3
done

if [ $SECONDS -ge 600 ]
then
    echo "Timed out waiting for ${NAME} to come up"
    exit 1
fi

