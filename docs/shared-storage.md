# Shared Storage Dependency and Configuration

## Overview

To support distributed job status management in Kubernetes deployments, `mxclone` uses a shared storage backend for job tracking. The recommended and default implementation is Redis, but the system is designed to be extensible for other backends (e.g., PostgreSQL, etcd) via the `JobStore` interface.

## Why Shared Storage?

- Ensures all application replicas (pods) can access and update job status consistently
- Enables reliable polling and progressive updates for async jobs (e.g., traceroute)
- Supports horizontal scaling in Kubernetes and other distributed environments

## Redis as Shared Job Store

- **Default Redis image:** `redis:alpine` (see `k8s/deployment.yaml`)
- **Service name:** `redis-service` (exposed on port 6379)
- **Key prefix:** `traceroutejob:` (configurable)
- **No authentication by default** (set password in config for production)

## Configuration

You can configure the shared storage via `config.yaml` or environment variables:

```yaml
job_store_type: "redis"
redis:
  redis_address: "redis-service:6379"
  redis_password: ""           # Set for production
  redis_db: 0
  redis_prefix: "traceroutejob:"
```

Or with environment variables (overrides config file):

```
MXCLONE_JOB_STORE_TYPE=redis
MXCLONE_REDIS_REDIS_ADDRESS=redis-service:6379
MXCLONE_REDIS_REDIS_PASSWORD=yourpassword
MXCLONE_REDIS_REDIS_DB=0
MXCLONE_REDIS_REDIS_PREFIX=traceroutejob:
```

## Kubernetes Deployment

- The `k8s/deployment.yaml` file includes both the `mxclone` app and a Redis container.
- The app is configured to use Redis for job storage by setting `MXCLONE_JOB_STORE_TYPE=redis`.
- The Redis service is exposed internally as `redis-service:6379`.
- For production, set a Redis password and update the deployment and config accordingly.

## Security Considerations

- **Do not expose Redis directly to the public internet.**
- Use Kubernetes secrets or environment variables for sensitive credentials.
- Restrict network access to Redis to only application pods.

## Extending JobStore

- The `JobStore` interface allows for alternative implementations (e.g., PostgreSQL, etcd).
- To add a new backend, implement the interface and update configuration as needed.

## References

- See `internal/config/config.go` for config structure
- See `internal/redis_job_store.go` for RedisJobStore implementation
- See `k8s/deployment.yaml` for deployment example
