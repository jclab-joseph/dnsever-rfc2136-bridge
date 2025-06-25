# dnsever-rfc2136-bridge

dnsever-rfc2136-bridge 은 DNSEver 의 DNS 기능을 RFC2136 프로토콜로 수정할 수 있게 해주는 프로그램 입니다.

주된 목적은 LetsEncrypt DNS-01 Challenge 을 자동으로 진행하기 위함입니다.

TXT 레코드만 지원합니다.

# 배포

## Helm

See [deploy/helm/README.md](./deploy/helm/README.md)

## authfile.yaml

```yaml
domains:
- zone: example.com         # Root Domain
  upstream:
    - ns303.dnsever.com:53  # DNSEver DNS Server
  clientId: "example"       # DNSEver ID
  clientSecret: "dnssecret" # 다이나믹DNS 인증코드 : https://kr.dnsever.com/myinfo.html?selected_menu=dnspreference
  tsig:
    - name: "example."
      secret: "E4UkMlWVBoEHfNic2tA2LsZqMpqcyi9fX/tw+lqkMgej7BwQk2RTi7VOS76UMQXt1AZEQNWstXyO5qS1FHABoQ==" # generate with "tsig-keygen -a hmac-sha512 example"
```

## certbot 사용

```bash
$ certbot certonly \
  --dns-rfc2136 \
  --dns-rfc2136-propagation-seconds 60 \
  --dns-rfc2136-credential rfc2136.ini \
  -d your.domain.example.com
```

**rfc2136.ini**:

```text
# Target DNS server
dns_rfc2136_server = 127.0.0.1
# Target DNS port
dns_rfc2136_port = 2053
# TSIG key name
dns_rfc2136_name = example
# TSIG key secret
dns_rfc2136_secret = E4UkMlWVBoEHfNic2tA2LsZqMpqcyi9fX/tw+lqkMgej7BwQk2RTi7VOS76UMQXt1AZEQNWstXyO5qS1FHABoQ==
# TSIG key algorithm
dns_rfc2136_algorithm = HMAC-SHA512
```

## cert-manager 예시


```bash
kubectl -n cert-manager create secret generic certmanager-dnsever-tsig --from-literal=tsig-secret=TSIG_SECRET
```


```yaml
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-dnsever-prod
spec:
  acme:
    email: YOUR@EMAIL.COM
    preferredChain: ""
    privateKeySecretRef:
      name: letsencrypt-dnsever-prod
    server: https://acme-v02.api.letsencrypt.org/directory
    solvers:
    - dns01:
        rfc2136:
          nameserver: DNSEVER_RFC2136_SERVER:2053
          tsigAlgorithm: HMACSHA512
          tsigKeyName: TSIG_KEY_NAME
          tsigSecretSecretRef:
            name: certmanager-dnsever-tsig
            key: tsig-secret
```

# License

[Apache 2.0 License](./LICENSE)
