# Setup hfrd on k8s with tls and basic authentication

1. Create a directory named hfrdsetup (or any name you prefer, using hfrdsetup
in this doc is to simplify the instruction) on your own machine where kubectl
is available
2. Download every file in this directory into hfrdsetup directory
3. Set your k8s environment with your kubeconfig file.  The search order for kubeconfig is as follows,
   a) Place your kubeconfig file and possibly certificate file in hfrdsetup
   directory
   b) If you already have an KUBECONFIG environment variable set   
4. Configure Helm/Tiller for your k8s cluster. If using IBM Cloud Kubernetes Services you can follow these instructions: https://cloud.ibm.com/docs/containers?topic=containers-helm
4. Run ./create.sh
5. Use the url displayed at the end of the run to access the service

This procedure stands up hfrd in your own k8s cluster and secure
it with self signed certificate and basic authentication. The default
userid is sammy, the password is ps.

If you like to use your own set of usernames,passwords and certificates
you will need a pair of certificates, and a set of username and passwords.
username and password must be generated using htpasswd tool.

Once you have your own certificates, username and passwords, modify
override.yaml file. 
1. use your username and password to replace the value of htpasswd element.
2. Use the private key and cert of your certificates to replace tls.key and tls.cert value. 

If you like to remove hfrd services from your k8s cluster, simply run the
delete.sh script.
