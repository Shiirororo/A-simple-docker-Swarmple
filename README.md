



TO QUERY SUM OF CPU USAGE ON APP CONTAINER


```
sum(irate(container_cpu_usage_seconds_total{container_label_com_docker_swarm_service_name="demo_app"}[1m]))
```


TO COUNT NUMBER OF CONTAINER:

```
count(container_last_seen{container_label_com_docker_swarm_service_name="demo_app"})
```