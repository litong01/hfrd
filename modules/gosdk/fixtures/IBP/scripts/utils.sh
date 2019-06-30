#!/bin/bash

# Parse yaml file: Based on https://gist.github.com/pkuczynski/8665367
function parse_yaml() {
    local yaml_file=$1
    local prefix=$2
    local s
    local w
    local fs

    s='[[:space:]]*'
    w='[a-zA-Z0-9_.-]*'
    fs="$(echo @|tr @ '\034')"

    (
        sed -ne '/^--/s|--||g; s|\"|\\\"|g; s/\s*$//g;' \
            -e "/#.*[\"\']/!s| #.*||g; /^#/s|#.*||g;" \
            -e  "s|^\($s\)\($w\)$s:$s\"\(.*\)\"$s\$|\1$fs\2$fs\3|p" \
            -e "s|^\($s\)\($w\)$s[:-]$s\(.*\)$s\$|\1$fs\2$fs\3|p" |

        awk -F"$fs" '{
            indent = length($1)/2;
            if (length($2) == 0) { conj[indent]="+";} else {conj[indent]="";}
            vname[indent] = $2;
            for (i in vname) {if (i > indent) {delete vname[i]}}
                if (length($3) > 0) {
                    vn=""; for (i=0; i<indent; i++) {vn=(vn)(vname[i])("_")}
                    printf("%s%s%s%s=(\"%s\")\n", "'"$prefix"'",vn, $2, conj[indent-1],$3);
                }
            }' |
            
        sed -e 's/_=/+=/g' |
        awk 'BEGIN {
                 FS="=";
                 OFS="="
             }
             /(-|\.).*=/ {
                 gsub("-|\\.", "_", $1)
             }
             { print }'

    ) < "$yaml_file"
}

function create_variables() {
    local yaml_file="$1"
    eval "$(parse_yaml "$yaml_file")"
}


# Gather the msp_ids
gatherMSPIDS(){
	profilePath=$1
	for profile in $( ls $profilePath )
	do
		if [[ $profile =~ 'ConnectionProfile_' ]]; then
			MSP_ID=${profile#*_}
			MSP_ID=${MSP_ID%.*}
			MSP_IDs=("${MSP_IDs[@]}" "$MSP_ID")
		fi
	done
}

get_pem() {
	awk '{printf "%s\\n", $0}' $DefaultCredsDir/"$1"admin/msp/signcerts/cert.pem
}

runWithRetry(){
	retrytime=0
	while [ $retrytime -le $MAX_RETRY ]; do
    	((retrytime++))
		# Use first parameter as the function
    	$1
    	if [ $? -ne 0 ]; then
        	if [ $MAX_RETRY -eq $retrytime ]; then
       			log "Already reach the maximum number of attempts.Still failed to finish the job"
				sleep 2s
        		break 1
    		else
            	log "Job failed,will retry"
				sleep 2s
        	fi
    	else
    		log "Job succeeded"
			sleep 2s
        	break 1
    	fi
	done
}