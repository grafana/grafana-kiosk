# Grafana Kiosk in Containers

If you would like to run Grafana Kiosk from a docker container, this is possible and I do so in my kubernetes cluster!

## Note

This only works for Raspberry Pi as it uses the Raspberry Pi OS as the base image. (See Debugging)

Your raspberry pi must be running X, newer versions run wayland by default so you have to change this:

* sudo raspi-config
* Advanced Options
* Wayland
* Select `W1 X11`
* OK -> Reboot

## Build

Pass the image name in as an argument, e.g.:

```bash
mage -v build:dockerArm64 "slimbean/grafana-kiosk:2024-11-29"
```

## Running in Kubernetes

Here's an example I use which creates a namespace, the deployment and also a cron job that restarts Grafana every day.

The cron job isn't really necessary, but I generally deploy things like this with some mechanism of restarting them to deal with any unexpected issues.

Also I have a few extra affinity and toleration settings, the node in my k3s cluster that I want to run this on has a label `role=grafana-kiosk` and also a taint so I can target it specifically.

```yaml
  affinity:
    nodeAffinity:
      requiredDuringSchedulingIgnoredDuringExecution:
        nodeSelectorTerms:
          - matchExpressions:
              - key: role
                operator: In
                values:
                  - grafana-kiosk
```

Also that same node has a taint `role=grafana-kiosk:NoSchedule` so I can ensure that only this pod runs on that node.

```yaml
tolerations:
  - effect: NoSchedule
    key: role
    operator: Equal
    value: grafana-kiosk
```

Toleration syntax is confusing to me, but what this says is that this pod "tolerates" the taint `role=grafana-kiosk` with the value `grafana-kiosk` and the effect `NoSchedule`.

Full YAML:

Make sure to update the URL, USR, PASS, and IMAGE fields.

```yaml
apiVersion: v1
kind: Namespace
metadata:
  creationTimestamp: null
  name: grafana-kiosk
spec: {}
status: {}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  creationTimestamp: null
  labels:
    app: grafana-kiosk
  name: grafana-kiosk
  namespace: grafana-kiosk
spec:
  replicas: 1
  selector:
    matchLabels:
      app: grafana-kiosk
  strategy:
    type: Recreate
  template:
    metadata:
      creationTimestamp: null
      labels:
        app: grafana-kiosk
        name: grafana-kiosk
    spec:
      affinity:
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
              - matchExpressions:
                  - key: role
                    operator: In
                    values:
                      - grafana-kiosk
      containers:
        - args:
            - -URL=FIXME URL
            - -login-method=local
            - -username=FIXME USER
            - -password=FIXME PASS
            - -kiosk-mode=full
          env:
            - name: DISPLAY
              value: unix:0.0
            - name: KIOSK_DEBUG
              value: "false"
          image: FIXME IMAGE
          imagePullPolicy: IfNotPresent
          name: grafana-kiosk
          ports:
            - containerPort: 3000
              name: http-metrics
              protocol: TCP
          resources:
            limits:
              cpu: "3"
              memory: 2000Mi
            requests:
              cpu: "1"
              memory: 500Mi
          volumeMounts:
            - mountPath: /tmp/.X11-unix
              name: x11
      tolerations:
        - effect: NoSchedule
          key: role
          operator: Equal
          value: grafana-kiosk
      volumes:
        - hostPath:
            path: /tmp/.X11-unix
          name: x11
status: {}
---
apiVersion: v1
kind: ServiceAccount
metadata:
  creationTimestamp: null
  name: kiosk-cron
  namespace: grafana-kiosk
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  creationTimestamp: null
  name: kiosk-cron
  namespace: grafana-kiosk
rules:
  - apiGroups:
      - apps
      - extensions
    resources:
      - deployments
    verbs:
      - get
      - patch
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  creationTimestamp: null
  name: kiosk-cron
  namespace: grafana-kiosk
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: kiosk-cron
subjects:
  - kind: ServiceAccount
    name: kiosk-cron
    namespace: grafana-kiosk
---
apiVersion: batch/v1
kind: CronJob
metadata:
  creationTimestamp: null
  name: kiosk-cron
  namespace: grafana-kiosk
spec:
  concurrencyPolicy: Forbid
  jobTemplate:
    metadata:
      creationTimestamp: null
    spec:
      backoffLimit: 2
      template:
        metadata:
          creationTimestamp: null
        spec:
          containers:
            - command:
                - kubectl
                - rollout
                - restart
                - deployment/grafana-kiosk
              image: rancher/kubectl:v1.22.2
              name: kubectl
              resources: {}
          restartPolicy: Never
          serviceAccountName: kiosk-cron
  schedule: 00 07 * * *
status: {}
```

## Debugging

Debugging chromium issues is tricky...

If you get a "context canceled" error when the kiosk is starting, this is typically because the chromium process failed to start or crashed.
You can enable `KIOSK_DEBUG=true` env var but this will only help so much.

Instead I modified the docker container to sleep forever so I could run chromium exec'd into it:

using:

```dockerfile
#ENTRYPOINT [ "/kiosk/grafana-kiosk" ]
CMD sleep infinity
```

Then inside the container I was trying to run what I believe to be the similar command as run by the kiosk program

```bash
chromium --autoplay-policy=no-user-gesture-required --bwsi --check-for-update-interval=31536000 --disable-atures=Translate --disable-notifications --disable-overlay-scrollbar --disable-sync --ignore-certificate-errors=false --incognito --kiosk --noerrdialogs --kiosk --start-fullscreen --start-maximized --user-agent="Mozilla/5.0 (X11; CrOS armv7l 13597.84.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36" --window-position="0,0" --user-data-dir="/tmp/123"
```

for debian bookworm it seemed like i made progress adding:

```bash
--no-zygote --no-sandbox
```

While disabling these options worked, I found it a bit unsatisfying... and googling led me to believe that this may somehow be a raspberry pi problem which was already handled by the raspberry pi os images.

Unfortunately they don't officially publish docker images for raspberry pi os, but I was able to find someone that is, using these images fixed my problems, even though they are huge images...

Still wish I had succeeded in getting the debian images to work, but maybe I'll revisit it again someday, and since I'm running on raspberry pi's and way exceeded how much time I want to spend on this, this is how it stays for now :)
