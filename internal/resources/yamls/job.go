package yamls

const DB_MIGRATION_MCSPID = `
apiVersion: batch/v1
kind: Job
metadata:
  name: account-iam-db-migration-mcspid
  namespace: mcsp
  labels:
    by-squad: mcsp-user-management
    for-product: all
    bcdr-candidate: t
    component-name: iam-services
  annotations:
    argocd.argoproj.io/sync-wave: "2"
    test: mcspid
spec:
  backoffLimit: 3
  activeDeadlineSeconds: 21600
  template:
    metadata:
      labels:
        by-squad: mcsp-user-management
        for-product: all
        name: account-iam
    spec:
      restartPolicy: Never
      containers:
        - name: dbmigrate
          image: icr.io/automation-saas-platform/access-management/account-iam:20240430132235-main-70f5d498c1867f5da50a2b1bf010eae8e4a4b1c1
          envFrom:
            - secretRef:
                name: account-iam-database-secret
          command:
            - /bin/sh
            - '-c'
          args:
            - '/dbmigration/run.sh'
          volumeMounts:
          imagePullPolicy: Always
          resources:
            requests:
              cpu: 100m
              memory: 300Mi
            limits:
              cpu: 500m
              memory: 600Mi
      restartPolicy: Never
      serviceAccountName: account-iam-migration
      volumes:
        - name: account-iam-token
          projected:
            sources:
              - serviceAccountToken:
                  audience: openshift
                  expirationSeconds: 7200
                  path: account-iam-token
            defaultMode: 420
`
