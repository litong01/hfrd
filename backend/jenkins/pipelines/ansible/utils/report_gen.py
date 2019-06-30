#!/usr/bin/python
import json,sys,os,yaml

reqId = sys.argv[1]
metrics_dir = sys.argv[2]

metrics_description = [
    '"Average TPS", ',
    '"Max transaction response time", ',
    '"Average transaction response time", ',
    '"Min transaction response time", ',
    '"Average transaction proposal time", ',
    '"Average transaction broadcast time", '
]

metrics_names = [
    "avg/sum(hfrd_chaincode_invoke_counter_ps{testid=\""+reqId+"\"})",
    "max/max(hfrd_chaincode_invoke_timer_max{testid=\""+reqId+"\"})",
    "avg/avg(hfrd_chaincode_invoke_timer_mean{testid=\""+reqId+"\"})",
    "min/min(hfrd_chaincode_invoke_timer_min{testid=\""+reqId+"\"})",
    "avg/avg(hfrd_chaincode_invoke_proposal_timer_mean{testid=\""+reqId+"\"})",
    "avg/avg(hfrd_chaincode_invoke_broadcast_timer_mean{testid=\""+reqId+"\"})"
]

summary_output = open(metrics_dir+"/"+reqId+"-metrics_summary.csv", "w")
summary_output.write('"Test Summary Report for test", ' + reqId +"\n")
#summary_output.write("Note: Time is in milliseconds." + "\n")


def generateAverageMetrics(metrics_name_index, metrics, summary_output):
    summary = 0.0
    avg_val = 0.0
    for index, item in enumerate(metrics):
        if len(metrics) < 3:
            summary += float(item[1])
        else:
            if index != 0 and index != len(metrics) - 1:
                summary += float(item[1])
    if len(metrics) > 0 and len(metrics) <= 2:
        avg_val = summary/len(metrics)
    elif len(metrics) > 2:
        avg_val = summary/(len(metrics)-2)
    summary_output.write(
        metrics_description[metrics_name_index] + str(avg_val) + "\n")


def generateMaxMetrics(metrics_name_index, metrics, summary_output):
    value_list = []
    for item in metrics:
        value_list.append(float(item[1]))
    summary_output.write(
        metrics_description[metrics_name_index] + str(max(value_list)) + "\n")


def generateMinMetrics(metrics_name_index, metrics, summary_output):
    value_list = []
    for item in metrics:
        value_list.append(float(item[1]))
    summary_output.write(
        metrics_description[metrics_name_index] + str(min(value_list)) + "\n")

# generate target TPS
try:
    testplan = open(metrics_dir+"/../testplan.yml", "r")
    ty = yaml.load(testplan, Loader=yaml.SafeLoader)
    targetTPSs = []
    currentTargetTPS = 0
    if ty['tests'] is None:
        raise Exception('No tests defined')
    for idx, test in enumerate(ty['tests']):
        try:
            if test['operation'] == 'CHAINCODE_INVOKE':
                if 'iterationInterval' not in test:
                    iterationInterval = '1s'
                else:
                    iterationInterval = test['iterationInterval']
                if iterationInterval.endswith('r') or iterationInterval.endswith('s'):
                    interval = float(iterationInterval[:-1])
                    if 'loadSpread' not in test:
                        loadSpread = 1
                    else:
                        loadSpread = test['loadSpread']
                    if interval > 0:
                        currentTargetTPS += loadSpread/interval
            if test['waitUntilFinish'] is not False or idx == len(ty['tests']) - 1:
                if currentTargetTPS > 0:
                    targetTPSs += [currentTargetTPS]
                    currentTargetTPS = 0
        except Exception as e:
            print 'iteration error:', e
    summary_output.write('"Target TPS", ' + str((','.join(map(str, targetTPSs)))) + '\n')
except Exception as e:
    print 'Generating target TPS as N/A:', e
    summary_output.write('"Target TPS", N/A\n',)

for metrics_name_index, metrics_name in enumerate(metrics_names):
    metrics = metrics_name.split("/")
    metrics_type = metrics[0]
    metrics_name = metrics[1]
    if os.path.exists(metrics_dir + "/" + metrics_name + ".json"):
        with open(metrics_dir + "/" + metrics_name + ".json", 'r') as f:
            try:
                data_loaded = json.load(f)
                if len(data_loaded["data"]["result"]) != 0:
                    metrics = data_loaded["data"]["result"][0]["values"]
                    if metrics_type == "max":
                        generateMaxMetrics(metrics_name_index, metrics, summary_output)
                    elif metrics_type == "avg":
                        generateAverageMetrics(metrics_name_index,metrics,summary_output)
                    elif metrics_type == "min":
                        generateMinMetrics(metrics_name_index, metrics, summary_output)
            except Exception as e:
                print "Error when load metrics:" + metrics_name, e

