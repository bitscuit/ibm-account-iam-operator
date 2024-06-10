# ibm-account-iam-operator
OpenShift operator to install and manage IBM MCSP account-iam service

## Developer guide
 
### Local development

1. Install the CRDs
   ```
   make install
   ```
1. Build the docker image of the binary
   ```
   make docker-build-dev
   ```
1. Install the deployment manifests
   ```
   make deploy
   ```
