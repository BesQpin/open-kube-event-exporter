# open-kube-event-exporter

`open-kube-event-exporter` is a lightweight, open-source **Kubernetes event exporter** written in **Go**.  
It listens to events across your Kubernetes cluster and exports them to **Prometheus** and/or **Loki**, allowing you to monitor and analyze event activity with your existing observability stack.

---

## 🚀 Features

- Exports Kubernetes events in real-time
- Supports both **Prometheus metrics** and **Loki logs**
- Built with the official **Kubernetes Go client**
- Minimal configuration and resource usage
- Works with **Helm**, **Flux**, or standalone `kubectl`
- Designed for modern Kubernetes environments

---

## 🧩 Architecture Overview

```text
┌───────────────────────────────┐
│   Kubernetes API Server       │
│   (cluster events)            │
└───────────────┬───────────────┘
                │
                ▼
     ┌──────────────────────────┐
     │ open-kube-event-exporter │
     │ (Go app)                 │
     │  ├─ Watches events       │
     │  ├─ Sends to Prometheus  │
     │  └─ Sends to Loki        │
     └────────────┬─────────────┘
                  │
   ┌──────────────┴──────────────┐
   │ Prometheus / Loki backend   │
   │ (Dashboards & alerts)       │
   └─────────────────────────────┘
```

---

## 🏗️ Installation

You can deploy the exporter directly with **Helm** or **Flux**.

### Option 1 — Helm (direct install)

1. **Authenticate to GitHub Container Registry (GHCR):**

   ```bash
   echo $CR_PAT | helm registry login ghcr.io --username <your-github-username> --password-stdin
   ```

2. **Install the chart:**

   ```bash
   helm install open-kube-event-exporter oci://ghcr.io/besqpin/open-kube-event-exporter/open-kube-event-exporter \
     --version v1.1.1 \
     --namespace dev \
     -f values.yaml
   ```

3. **Verify the deployment:**

   ```bash
   kubectl get pods -n dev -l app=open-kube-event-exporter
   ```

---

### Option 2 — Flux (GitOps deployment)

If you are using Flux, reference the chart as an OCI Helm source:

```yaml
apiVersion: source.toolkit.fluxcd.io/v1
kind: HelmRepository
metadata:
  name: open-kube-event-exporter
  namespace: flux-system
spec:
  type: oci
  url: oci://ghcr.io/besqpin/open-kube-event-exporter
---
apiVersion: helm.toolkit.fluxcd.io/v2
kind: HelmRelease
metadata:
  name: open-kube-event-exporter
  namespace: monitoring
spec:
  interval: 5m
  chart:
    spec:
      chart: open-kube-event-exporter
      sourceRef:
        kind: HelmRepository
        name: open-kube-event-exporter
      version: ">=1.1.1"
  values:
    loki:
      url: "https://loki-gateway.loki.svc.cluster.local/loki/api/v1/push"
      tenantID: "default"
    prometheus:
      enabled: true
```

---

## ⚙️ Configuration

### Example `values.yaml`

```yaml
replicaCount: 1

image:
  repository: ghcr.io/besqpin/open-kube-event-exporter
  tag: "v1.1.1"
  pullPolicy: IfNotPresent

service:
  type: ClusterIP
  port: 8080

loki:
  url: "https://loki-gateway.loki.svc.cluster.local/loki/api/v1/push"
  tenantID: "default"

prometheus:
  enabled: true
  serviceMonitor:
    interval: 30s
    scrapeTimeout: 10s
```

---

## 📊 Metrics

When Prometheus is enabled, metrics are exposed at:

```
http://<pod-ip>:8080/metrics
```

Example metric output:

```
# HELP kube_event_total Total number of Kubernetes events
# TYPE kube_event_total counter
kube_event_total{namespace="default",reason="Created",type="Normal"} 42
```

---

## 🪵 Loki Integration

If Loki is configured, events are pushed as structured JSON log entries.

Example log entry:

```json
{
  "namespace": "default",
  "name": "nginx-deployment-5d8b6b9d6f-abcde",
  "reason": "Scheduled",
  "message": "Successfully assigned pod to node aks-nodepool-1",
  "type": "Normal",
  "timestamp": "2025-10-10T12:34:56Z"
}
```

---

## 🧠 Development

### Requirements

- Go 1.23 or later
- Docker
- Helm 3.9+ (for packaging)
- kubectl with access to a cluster

### Run locally

```bash
go mod tidy
go run main.go
```

---

## 🐳 Build & Push Docker Image

```bash
docker build -t ghcr.io/<your-org>/open-kube-event-exporter:latest .
docker push ghcr.io/<your-org>/open-kube-event-exporter:latest
```

---

## 🧩 Helm Chart Packaging (OCI)

```bash
helm package charts/open-kube-event-exporter
helm push open-kube-event-exporter-*.tgz oci://ghcr.io/<your-org>/open-kube-event-exporter
```

---

## 📜 License

This project is licensed under the **MIT License**.  
See the [LICENSE](./LICENSE) file for details.

---

## 💡 Maintainer

**Peter Williams**  
GitHub: [@BesQpin](https://github.com/BesQpin)

---

> Open, simple, and built for the Kubernetes community 🚀
