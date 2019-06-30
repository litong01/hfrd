#!/bin/bash -x


CACERT=$(base64 ${CACERT} | tr -d '\n')
ADMINCERT=$(base64 ${ADMINCERT} | tr -d '\n')

cat << EOB > ${LOCATION}
{
	"enrollment": {
		"component": {
			"cahost": "${CA_HOST}",
			"caport": "${CA_PORT}",
			"caname": "${CA_NAME}",
			"catls": {
				"cacert": "${CACERT}"
			},
			"enrollid": "${ENROLL_ID}",
			"enrollsecret": "${ENROLL_SECRET}",
			"admincerts": ["${ADMINCERT}"]
		},
		"tls": {
			"cahost": "${CA_HOST}",
			"caport": "${CA_PORT}",
			"caname": "${TLSCA_NAME}",
			"catls": {
				"cacert": "${CACERT}"
			},
			"enrollid": "${ENROLL_ID}",
			"enrollsecret": "${ENROLL_SECRET}",
			"csr": {
				"hosts": [
					${CSRHOSTS}
				]
			}
		}
	}
}
EOB

kubectl create secret generic ${NAME} --from-file=${LOCATION} --from-literal=couchdbusr=couchdb_user --from-literal=couchdbpwd=couchdb_pass