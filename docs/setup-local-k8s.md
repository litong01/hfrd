# Set up local k8s cluster(single node)
In **HFRD** , k8s cluster is used to **set up fabric network** using cello and **run module tests** against fabric network. We can get a k8s cluster from `cloud` or `set up by our own`. This document will show you how to set up a local k8s cluster which can be used in **HFRD**.
If you test is agains cloud environment which means your fabric network can be accessed by public IP address,you can use both `cloud k8s cluster` and `local k8s cluster` for testing.But if your fabric network can only be accessed in private ip, then you must set up your own `k8s ` cluster. 

First of all, you must have a linux machine (x86 or s390x). Following processes have be verified in linux vm(ubuntu 16.04) provided by IBM vLauncher.(**https://vlaunch.rtp.raleigh.ibm.com/**)

#### Prequsities
* Docker (latest docker-ce)			
	`apt-get install docker.io`
* Git (Configure git to clone **github.ibm.com** projects)

#### Clone hfrd-nfs-provisioner
```
	git clone git@github.ibm.com:bjwswang/hfrd-local-kubernetes.git
	cd hfrd-local-kubernetes
```

#### Set up single node k8s cluster
```
	sudo su
	cd k8s-setup/
	chmod +x *
	./setup_k8s_{OS_ARCH}.sh
```
#### Start nfs provisioner
```
	sudo su
	cd ..
	cd nfs-provisioner/
	chmod +x *
	./setup_nfs_provisioiner.sh
```
#### Package `kubeconfig` into zip file(no-root user)
After k8s cluster is created , we need to package `kubeconfig` file for hfrd.

```
	cd /home/ibmadmin
	mkdir kubeconfig 
	sudo cp /root/.kube/config kubeconfig/kubeconfig.yml
	sudo chown -R ibmadmin:ibmadmin kubeconfig/kubeconfig.yml
	zip -r kubeconfig.zip kubeconfig/
```





