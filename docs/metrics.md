## Backend monitoring

Для backend теперь настроен стек мониторинга:

- `Prometheus`: `http://localhost:9090`
- `Grafana`: `http://localhost:3001`
- Grafana dev login: `admin`
- Grafana dev password: `admin`
- Grafana credentials можно переопределить через `GRAFANA_ADMIN_USER` и `GRAFANA_ADMIN_PASSWORD`

### Как запустить

```bash
docker compose -f deployments/dev/compose.yaml up --build
```

Для Docker Desktop `node-exporter` запускается в совместимом read-only режиме через `/host`.

После старта должны быть `UP` targets:

- `api-gateway`
- `auth-service`
- `user-service`
- `movie-service`
- `node-exporter`
- `cadvisor`
- `prometheus`

### Где лежат конфиги мониторинга

- Prometheus config: `deployments/monitoring/prometheus/prometheus.yml`
- Grafana datasource provisioning: `deployments/monitoring/grafana/provisioning/datasources/prometheus.yml`
- Grafana dashboard provisioning: `deployments/monitoring/grafana/provisioning/dashboards/dashboards.yml`
- Dashboard JSON: `deployments/monitoring/grafana/dashboards/vkino-overview.json`

### Что смотреть в Grafana

Dashboard: `Vkino Microservices Overview`

На нем собраны:

- HTTP hits / errors / timings для `api-gateway`
- gRPC hits / errors / timings для `auth-service`, `user-service`, `movie-service`
- host CPU / memory / disk
- container CPU / memory
