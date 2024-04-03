# Buildkit ConsistentHash
- Chart: [bradfordwagner/chart-docker-buildkit](https://github.com/bradfordwagner/chart-docker-buildkit)
- Upstream Server can be found here at [bradfordwagner/go-cli-buildkit-ch](https://github.com/bradfordwagner/go-cli-buildkit-ch)
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
