# ibm-user-management-operator
OpenShift operator to install and manage IBM MCSP account-iam service

## Developer guide
 
### Local development

1. Install the CRDs
   ```sh
   make install
   ```
2. Install the deployment manifests
   ```sh
   make deploy
   ```
3. Build the docker image of the binary
   - In `ibm.mk`, set the development image tag with the version or string (e.g., "dev", "latest",...), and the development registry 
     ```sh
     DEV_VERSION ?=1.0.0
     DEV_REGISTRY ?= quay.io/bedrockinstallerfid
     ```
   - Build the image
     ```sh
     make docker-build-dev
     ```
4. Build and push bundle and catalog images
   ```sh
   make bundle-build-push-dev
   make catalog-build-push-dev
   ```