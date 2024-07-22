package yamls

const DB_BOOTSTRAP_JOB = `
apiVersion: batch/v1
kind: Job
metadata:
  name: create-account-iam-db
spec:
  template:
    metadata:
      name: create-account-iam-db
    spec:
      containers:
      - name: postgres
        image: docker-na-public.artifactory.swg-devops.com/hyc-cloud-private-integration-docker-local/ibmcom/mcsp-utils:latest
        command: ["/bin/bash", "/db-init/create_db.sh"]
        volumeMounts:
        - name: psql-credentials
          mountPath: /psql-credentials
        - name: db-password
          mountPath: /db-password
        - name: data-volume
          mountPath: /data
      restartPolicy: OnFailure
      volumes:
      - name: psql-credentials
        secret:
          secretName: common-service-db-superuser
          items:
          - key: username
            path: username
          - key: password
            path: password
          defaultMode: 420  
      - name: db-password
        secret:
          secretName: account-im-db-password
          items:
          - key: password
            path: password
          defaultMode: 420  
      - name: data-volume
        emptyDir: {}      
  backoffLimit: 4

`
