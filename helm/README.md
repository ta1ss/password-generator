# Deploying in K8s

```bash
$ kubectl create namespace password-generator
```

### Add your wordlist in the configMap
```bash
$ kubectl -n password-generator create configmap my-wordlist --from-file=wordlist.
txt
```

### Adjust the values in the values.yaml: 
```yaml
valuesConfigMap:
  enabled: true
  data:
    values.yaml: |
      MIN_PASSWORD_LENGTH: 15           # minimum length of the password
      MAX_PASSWORD_LENGTH: 32           # maximum length of the password
      BETWEEN_SYMBOLS: ""               # define symbols for between the words
      INSIDE_SYMBOLS: ""                # define symbols for the words
      PASSWORD_PER_ROUTINE: 300         # generated passwords per GO
      routine          
      SYMBOL_MAPPING:                   # define which char you want to be swapped     
        # key: value                    # value is mapped to key
        # a: b
      WORDLIST_PATH: "data/wordlist.txt"  # Path to wordlist    

```

### Deploy: 
```bash
helm install password-generator . -n password-generator -f values.yaml
```

### Upgrade:
```bash
helm upgrade password-generator . -n password-generator -f values.yaml
```

---

### WoodpeckerCI & ArgoCD
You can utilize the premade CI/CD pipeline configurations through connecting `password-generator` with [WoodpeckerCI](https://woodpecker-ci.org/docs/intro) and [ArgoCD](https://argo-cd.readthedocs.io/en/stable/) (including [image-updater](https://argocd-image-updater.readthedocs.io/en/stable/)). 