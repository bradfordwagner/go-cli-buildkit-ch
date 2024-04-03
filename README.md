# Buildkit ConsistentHash
- This component is part of the [Buildkit](https://github.com/bradfordwagner/chart-docker-buildkit) to add a service discovery component based on consistent hashing to tell the client which buildkit instance to send the request to.
- Downstream container can be found here at [bradfordwagner/container-bkch](https://github.com/bradfordwagner/container-bkch)

## Usage
```bash
# in cluster
hash=abcdef
curl localhost:8888/in-cluster/${hash}

# outputs
buildkit-0.buildkit.buildkit.svc.cluster.local

# using api gateway
hash=abcdef
curl localhost:8888/api-gateway/${hash}

# outputs
buildkit-0.mydomain.com
```
