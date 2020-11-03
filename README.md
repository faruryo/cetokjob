# cetokjob

![Docker](https://github.com/faruryo/cetokjob/workflows/Docker/badge.svg)

Receive Events in the form of CloudEvents and generate a job with the data packed into environment variables.

## Usage

TODO

## Developing

### skaffold

```shell
skaffold dev --port-forward
```

### curl sample for firing CloudEvents

``` test
curl -v "localhost:8080" \
    -X POST \
    -H "Ce-Id: 536808d3-88be-4077-9d7a-a3f162705f79" \
    -H "Ce-Specversion: 1.0" \
    -H "Ce-Type: sample" \
    -H "Ce-Source: sample" \
    -H "Content-Type: application/json" \
    -d '{"msg":"Hello World from the curl pod."}'
```

## Tips

### Command to remove all jobs

```
kubectl get job -o name | xargs kubectl delete
```