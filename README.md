# Firefly

// todo: What's Firefly?

## Architecture

// TBD

## Quick Start

This guide will cover:

- Install firefly components in a Kubernetes cluster which is known as host cluster.
- Create a karmada instance and join a member cluster to it.

### Install firefly

Before you do it, please make sure that the [cert-manager](https://cert-manager.io) has been installed on the host cluster. 

**Step 1:** Install CRDs

```console
kubectl apply -f https://raw.githubusercontent.com/carlory/firefly/main/deploy/install.firefly.io_karmadas.yaml
```

**Step 2:** Create namespace

```console
kubectl create ns firefly-system
```

**Step 3:** Install components

```console
kubectl apply -f https://raw.githubusercontent.com/carlory/firefly/main/deploy/install.yaml
```

### Create a Karmada instance and join a member cluster to it

**Step 1:** Create a Karmada instance named demo

You can install karmada on any namespace. The given command will deploy karmada commponents on 
the `firefly-system` namespace.

```console
kubectl apply -f https://raw.githubusercontent.com/carlory/firefly/main/example/demo.yaml
```

Waiting for a few seconds, the output of `kubectl -n firefly-system get pods,services,secrets`:

```console
NAME                                                         READY   STATUS    RESTARTS   AGE
pod/demo-etcd-0                                             1/1     Running   0          21s
pod/demo-firefly-karmada-manager-5678598d95-dcrf5           1/1     Running   0          8s
pod/demo-karmada-aggregated-apiserver-65c77fbf4d-vdr9k      1/1     Running   0          11s
pod/demo-karmada-apiserver-8495c66b47-kllwn                 1/1     Running   0          21s
pod/demo-karmada-controller-manager-6fbd48544b-jdd7f        1/1     Running   0          8s
pod/demo-karmada-kube-controller-manager-57f9fd76f6-z9whl   1/1     Running   0          8s
pod/demo-karmada-scheduler-64b65b45d8-sxw8q                 1/1     Running   0          8s
pod/demo-karmada-webhook-5488959847-n7lzk                   1/1     Running   0          8s
pod/firefly-controller-manager-7fd49597b6-4s4zz              1/1     Running   0          69s

NAME                                         TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)             AGE
service/demo-etcd                           ClusterIP   None             <none>        2379/TCP,2380/TCP   21s
service/demo-karmada-aggregated-apiserver   ClusterIP   10.111.178.231   <none>        443/TCP             11s
service/demo-karmada-apiserver              ClusterIP   10.111.66.86     <none>        5443/TCP            21s
service/demo-karmada-webhook                ClusterIP   10.100.216.76    <none>        443/TCP             8s

NAME                                               TYPE                                  DATA   AGE
secret/default-token-4chv4                         kubernetes.io/service-account-token   3      74m
secret/demo-etcd-cert                             Opaque                                4      21s
secret/demo-firefly-karmada-manager-token-fpptr   kubernetes.io/service-account-token   3      42m
secret/demo-karmada-cert                          Opaque                                16     21s
secret/demo-karmada-webhook-cert                  Opaque                                2      21s
secret/demo-kubeconfig                            Opaque                                1      21s
secret/firefly-controller-manager-token-w6xzh      kubernetes.io/service-account-token   3      69s
```

**Step 2:** Join an existing cluster to the karmada using the `karmadactl` binary

Before doing it, we have to configure the hosts resolve if the host machine cannot directly access to the cluster domain of the `demo-karmada-apiserver` service. Because this karmada instance isn't exposed when we create it. 

Note that `10.111.66.86` is the clusterIP of the `demo-karmada-apiserver` service.

```console
echo '10.111.66.86    demo-karmada-apiserver.firefly-system.svc.cluster.local' >> /etc/hosts
```

Now, let's join an exising cluster `k8s` to the karmada instance.

```console
kubectl get  -n firefly-system secret demo-kubeconfig -ojsonpath='{.data.kubeconfig}' | base64 -d > config
karmadactl join ik8s --kubeconfig config --cluster-kubeconfig <your_cluster_kubeconfig> --cluster-context  <your_cluster_context>
```

After a member cluster is added, the correponding `scheduler-estimator` component will be auto installed by the `firefly-karamda-manager` component.

```console
(⎈ |ik8s01:firefly-system)➜  ~ kubectl -n firefly-system get po,svc,rolebinding
NAME                                                         READY   STATUS    RESTARTS   AGE
pod/demo-etcd-0                                             1/1     Running   0          2m22s
pod/demo-firefly-karmada-manager-5678598d95-dcrf5           1/1     Running   0          2m9s
pod/demo-karmada-aggregated-apiserver-65c77fbf4d-vdr9k      1/1     Running   0          2m12s
pod/demo-karmada-apiserver-8495c66b47-kllwn                 1/1     Running   0          2m22s
pod/demo-karmada-controller-manager-6fbd48544b-jdd7f        1/1     Running   0          2m9s
pod/demo-karmada-kube-controller-manager-57f9fd76f6-z9whl   1/1     Running   0          2m9s
pod/demo-karmada-scheduler-64b65b45d8-sxw8q                 1/1     Running   0          2m9s
pod/demo-karmada-webhook-5488959847-n7lzk                   1/1     Running   0          2m9s
pod/firefly-controller-manager-7fd49597b6-4s4zz              1/1     Running   0          3m10s
pod/karmada-scheduler-estimator-ik8s-69c8656785-r7lzq        1/1     Running   0          17s

NAME                                         TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)             AGE
service/demo-etcd                           ClusterIP   None             <none>        2379/TCP,2380/TCP   2m23s
service/demo-karmada-aggregated-apiserver   ClusterIP   10.111.178.231   <none>        443/TCP             2m13s
service/demo-karmada-apiserver              ClusterIP   10.111.66.86     <none>        5443/TCP            2m23s
service/demo-karmada-webhook                ClusterIP   10.100.216.76    <none>        443/TCP             2m10s
service/karmada-scheduler-estimator-ik8s     ClusterIP   10.103.181.159   <none>        10352/TCP           18s

NAME                                                                  ROLE                AGE
rolebinding.rbac.authorization.k8s.io/demo-firefly-karmada-manager   ClusterRole/admin   2m10s
```

## What's Next

See [RoadMap](ROADMAP.md) for details.

More will be coming Soon. Welcome to [open an issue](https://github.com/carlory/firefly/issues) and [propose a PR](https://github.com/carlory/firefly/pulls). 🎉🎉🎉

## Contributors

<a href="https://github.com/carlory/firefly/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=carlory/firefly" />
</a>

Made with [contrib.rocks](https://contrib.rocks).

## License

Firefly is under the Apache 2.0 license. See the [LICENSE](LICENSE) file for details.