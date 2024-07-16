package yamls

var IMConfigJob = `
apiVersion: batch/v1
kind: Job
metadata:
  name: mcsp-im-config-job
  labels:
    app: mcsp-im-config-job
spec:
  template:
    metadata:
      labels:
        app: mcsp-im-config-job
    spec:
      containers:
      - name: mcsp-im-config-job
        image: docker-na-public.artifactory.swg-devops.com/hyc-cloud-private-integration-docker-local/ibmcom/mcsp-utils:latest
        command: ["./mcsp-im-config-job"]
        imagePullPolicy: Always
        securityContext:
          allowPrivilegeEscalation: false
          capabilities:
            drop:
              - ALL
          runAsNonRoot: true
          seccompProfile:
            type: RuntimeDefault
        env:
          - name: LOG_LEVEL
            value: debug
          - name: CS_NAMESPACE
            value: {{ .Namespace }}
          - name: IM_HOST_URL
            value: {{ .IMHostURL }}
          - name: ACCOUNT_IAM_URL
            value: {{ .AccountIAMURL }}
      serviceAccountName: mcsp-utils
      restartPolicy: OnFailure
`
