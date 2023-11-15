# HELM chart


### Installing 
```
kubectl create namespace password-generator
helm install password-generator helm/ -n password-generator
kubectl apply -f helm/deployment.yaml -n password-generator
```

### Upgrading

```
helm upgrade password-generator helm/ --values helm/values.yaml -n password-generator
```