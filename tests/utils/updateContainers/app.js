const request = require('superagent');

const OIDC_ENDPOINTT = 'https://iam.stage1.ng.bluemix.net/oidc/token';
let payload = 'grant_type=urn%3Aibm%3Aparams%3Aoauth%3Agrant-type%3Aapikey&apikey=';

const yargs = require('yargs')
// user input
let argv = yargs.usage('Usage: node $0 --cmBaseUrl [url] --networkId [networkId] --apiKey [apiKey] --nodeType ' +
                       '[nodeType] --cpu [num] --memory [num]')
                .demandOption(['cmBaseUrl', 'networkId', 'apiKey', 'nodeType'])
                .example('node $0 --cmBaseUrl https://cm-ash-perf.4.secure.blockchain.ibm.com:444 ' +
                        '--networkId 31f53963e39342738d3804a2d0470db5 ' +
                        '--apiKey OIDXvGPNea41CsXYEpn37GvoNHMib5unuQUtpWpps8hC ' +
                        '--nodeType peer --cpu 1.2 --memory 1024')
                .argv
if ((!argv.cpu || argv.cpu === true) && (!argv.memory || argv.memory === true)) {
    console.error('need to provide at least one of cpu/memory number to be set')
    process.exit(1)
}
let cmBaseUrl = argv.cmBaseUrl;
let networkId = argv.networkId;
let apiKey = argv.apiKey;
let nodeType = argv.nodeType;
let cpu = argv.cpu;
let memory = argv.memory;

// cluster manager related
const listContainersUrl = cmBaseUrl + '/api/manager/listContainers';
const updateContainerUrl = cmBaseUrl + '/api/manager/updateContainer'

// index of each element in a container row
const ROW_KEYS = {
    NODE_NAME: 0,
    NODE_ID: 1,
    OWNER: 2,
    NODE_TYPE: 3,
    STATUS: 4,
    IP: 5,
    CPU: 6,
    MEMORY: 7
}

// return a Promise which contains access token
async function getAccessToken(oidcEndpoint) {
    try {
        let res = await request.post(oidcEndpoint).set('Accept', 'application/json')
            .set('Content-type', 'application/x-www-form-urlencoded')
            .send(payload + apiKey);
        if (res.statusCode != 200 || !res.body
            || !res.body.access_token) {
            throw new Error('no access_token in response body')
        }
        return res.body.access_token
    } catch(err) {
        throw new Error('unable to get access token from ' + oidcEndpoint + ': ' + err.message)
    }
}

// list containers given network id
async function listContainers(networkId) {
    let accessToken = await getAccessToken(OIDC_ENDPOINTT);
    let res = await request.get(listContainersUrl + '/' + networkId)
        .set('Authorization', accessToken)
    if (res.statusCode !== 200) {
        throw new Error('network with id:' + networkId + " not found!" + res.statusCode + '' + res.text)
    }
    let containersTable = res.text;
    let containerRawRows = containersTable.split('\n');
    let re = new RegExp(/\[(.*)\]/);
    let containerRows = [];
    for (let rawRow of containerRawRows) {
        let r = rawRow.match(re);
        if (r) {
            let row = r[1].replace(/<.*?>/g, '').replace(/\"/g, '').split(',')
            row.splice(8)
            if (row.length === 8) {
                containerRows.push(row)
            }
        }
    }
    if (containerRows.length < 1) {
        throw new Error("No containers found with network id:" + networkId)
    }
    return containerRows
}

// update cpu and memory of containers with node id
async function updateContainer(nodeId, networkId, cpu, memory) {
    let accessToken = await getAccessToken(OIDC_ENDPOINTT);
    let req = request.post(updateContainerUrl + '/' + nodeId + '/' + networkId)
        .set('Authorization', accessToken)
        .type('form')
    if (cpu) {
        req.send('cpu=' + cpu)
    }
    if (memory) {
        req.send('memory=' + memory)
    }
    return req
}

(async function main(){
    try {
        let containerRows = await listContainers(networkId)
        // Now each element in containerRows array is also an array whose length is 8
        // the 8 elements in each containerRow is AS ROW_KEYS
        for (let row of containerRows) {
            if (row[ROW_KEYS.STATUS] !== 'DELETED' && row[ROW_KEYS.NODE_TYPE].includes(nodeType)) {
                if ((!cpu || row[ROW_KEYS.CPU] == cpu) && (!memory || row[ROW_KEYS.MEMORY] == memory + ' MB')) {
                    console.log(row[ROW_KEYS.NODE_NAME], 'cpu:', row[ROW_KEYS.CPU], 'memory:', row[ROW_KEYS.MEMORY],
                        'no need to change cpu and memory')
                    continue
                }
                let res = await updateContainer(row[ROW_KEYS.NODE_ID], networkId, cpu, memory)
                if (res.statusCode !== 200 || !res.body || !res.body[0] || !res.body[0].result ||
                    !res.body[0].result.includes('Success')) {
                    console.error('update cpu/memory failure for', networkId, row[ROW_KEYS.NODE_TYPE],
                        row[ROW_KEYS.NODE_ID], row[ROW_KEYS.CPU], row[ROW_KEYS.MEMORY],
                        'code:', res.statusCode, 'text:', res.text)
                } else {
                    console.log('successfully changed cpu/memory for', row[ROW_KEYS.NODE_NAME])
                }
            }
        }

        // after the cpu/memory change, list the containers with new cpu/memory configuration
        console.log('containers list of type ' + nodeType + ' after cpu/memory change')
        containerRows = await listContainers(networkId)
        for (let row of containerRows) {
            if (row[ROW_KEYS.STATUS] !== 'DELETED' && row[ROW_KEYS.NODE_TYPE].includes(nodeType)) {
                console.log(row[ROW_KEYS.NODE_NAME], '|', row[ROW_KEYS.CPU], '|', row[ROW_KEYS.MEMORY])
            }
        }
    } catch(err) {
        console.error('caught error: ', err.message)
        process.exit(1)
    } finally {
        // clean up here
    }
})()
