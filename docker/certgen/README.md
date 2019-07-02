# The cert package generation tool
This tool is to generate cert file package

# How to use it?

1. Download the tool container image

```
   docker pull hfrd/certgen:latest
```

2. Create a working directory, name it allcerts for example
3. Place your kubeconfig file in allcerts directory
4. Download all your msp json files from your wallet
and also place them in the allcerts directory
5. Change to allcerts directory and run the following command:
```
   docker run -v $(pwd):/opt/agent/vars --rm hfrd/certgen:latest \
   ansible-playbook -e "namespace=default" certgen.yml
```
6. If the command finishes successfully, you should have a file
named certs.tgz in the allcerts directory.

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
