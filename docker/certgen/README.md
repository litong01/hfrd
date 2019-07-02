# The cert package generation tool
This tool is to generate cert file package

# How to use it?

1. Download the tool container image
```bash
docker pull hfrd/certgen:latest
```
2. Create a working directory, name it *allcerts* for example or
anything you like
3. Place your kubeconfig file in *allcerts* directory
4. Download all your msp json files from your wallet
and also place them in the *allcerts* directory
5. Change to *allcerts* directory and run the following command:

```bash
docker run -v $(pwd):/opt/agent/vars --rm hfrd/certgen:latest \
ansible-playbook -e "namespace=default" certgen.yml
```
6. If the command finishes successfully, you should have a file
named certs.tgz in the *allcerts* directory.

Replace the word default in the command with your actual k8s
namespace if your fabric network is not in the default namespace

# IMPORTANT:
When you create fabric network:
1. When you create your MSP via the console, it is very important to use
same id for your msp name, msp id and admin identity. For example, for
a new organization, when you create its msp, you can use word `myneworg`
for msp name, msp id, and msp admin id. Without doing this, the tool
may generate incorrect certs or may not find the right information
2. When download your MSP json file, make sure that you are using the
same name (normally you do not have to change it) as the msp id.
3. The names of peer, orderer should only use numeric numbers and lower
case characters. Character dash - and underscore _ can not be used.

# EXTREMELY IMPORTANT:
Here are the steps that you have to do in IBP Console manually:
1. Create an Certificate Authority if you do not already have one.
2. Create 1 msp per organization. Export Identity and save each to
a directory where you will need it to run the certgen tool.
3. Create an orderer service with either 1 node or 5 nodes using the
msp created in step 2 for orderer org.
4. Add all the orgs into the Consortium.
5. Create peers for each peer org created in step 2.

Once all the steps are complete, you can run the certgen tool by following
the steps described in how to use it section
