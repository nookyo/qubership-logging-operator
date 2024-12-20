This guide contains information of how to configure TLS for Logging components.

# Table of Content

* [Table of Content](#table-of-content)
* [Steps to renew certificates](#steps-to-renew-certificates)
* [Self-signed certificate generation](#self-signed-certificate-generation)
* [Logging deploy schemes](#logging-deploy-schemes)
  * [Graylog on VM](#graylog-on-vm)
  * [Graylog in Cloud](#graylog-in-cloud)
  * [Configure TLS for Logging](#configure-tls-for-logging)
    * [Configure TLS for Graylog UI on VM](#configure-tls-for-graylog-ui-on-vm)
    * [Configure TLS for Graylog Inputs on VM](#configure-tls-for-graylog-inputs-on-vm)
    * [Configure TLS for Graylog Inputs into Cloud](#configure-tls-for-graylog-inputs-into-cloud)
    * [Configure TLS for Graylog HTTP interface into Cloud](#configure-tls-for-graylog-http-interface-into-cloud)
    * [Configure TLS for Fluentd output](#configure-tls-for-fluentd-output)
    * [Configure TLS for FluentBit output](#configure-tls-for-fluentbit-output)
    * [About cert-manager](#about-cert-manager)
      * [Logging-operator integration with cert-manager](#logging-operator-integration-with-cert-manager)
* [Examples](#examples)
  * [Graylog in cloud with manually created secrets](#graylog-in-cloud-with-manually-created-secrets)
  * [Graylog in cloud with cert-manager integration](#graylog-in-cloud-with-cert-manager-integration)
  * [Fluentd in cloud with manually created secrets](#fluentd-in-cloud-with-manually-created-secrets)
  * [Fluentd in cloud with cert-manager integration](#fluentd-in-cloud-with-cert-manager-integration)
  * [Fluent-bit in cloud with manually created secrets](#fluent-bit-in-cloud-with-manually-created-secrets)
  * [Fluent-bit in cloud with cert-manager integration](#fluent-bit-in-cloud-with-cert-manager-integration)

# Steps to renew certificates

TLS certificates have a limited duration, and you can update them over time by replacing the content of
corresponding Secrets. Cert-manager can do that automatically if you set `renewBefore` parameter for certificate.

Unfortunately, in some cases, the renewal of certificates at destinations doesn't happen automatically and additional
steps are required. The table below reflects actions required after updating the certificates:

| Service             | No extra steps | Restart pod |
|---------------------|----------------|-------------|
| Graylog HTTP server |                | ✅           |
| Graylog inputs      |                | ✅           |
| Fluentd             |                | ✅           |
| Fluent-bit          |                | ✅           |

After restart, components may be unavailable for a short time.

# Self-signed certificate generation

You must have a correct certificate to use it for TLS: it has to contain valid IPs and DNS names as alt names,
and it must not be expired.

In case when you use a wrong certificate, you can face the following problem: log in to Graylog UI and go to
`System -> Nodes`. If you use an incorrect certificate, you'll see the message
`System information is currently unavailable` instead of nodes' information.

If you want to use self-signed certificate, you can create it yourself, for example by using
[OpenSSL](https://www.openssl.org/) tool.

You can use the following SSL config file:

```ini
[ req ]
default_bits           = 2048
default_keyfile        = graylog.key
distinguished_name     = req_distinguished_name
req_extensions         = v3_req
prompt                 = no

[ req_distinguished_name ]
C                      = US
ST                     = Test ST
L                      = Test L
O                      = Qubership
OU                     = Operations Platform
CN                     = graylog
emailAddress           = test@email.address

[ v3_req ]
basicConstraints = critical, CA:FALSE
keyUsage = critical, digitalSignature, keyEncipherment
extendedKeyUsage = clientAuth, serverAuth
subjectAltName = @alt_names

[alt_names]
DNS.1 = graylog-service.logging.svc
DNS.2 = graylog-service.logging-operator.svc
IP.1 = x.x.x.x
IP.2 = x.x.y.y
```

Pay attention to the last section that must contain each IP and DNS name that you're going to use.

Then you can use the commands below to generate a private key in the PKCS#8 format and a self-signed certificate.

Firstly, you have to create file the SSL config file. You can use the example above as a template. Let the file be
called `graylog.cnf`.

Secondly, you must generate RSA private key (`graylog.key`):

```bash
openssl genrsa -out graylog.key 4096
```

Then create certificate request (`graylog.csr`):

```bash
openssl req -new -out graylog.csr -key graylog.key -config graylog.cnf -nodes
```

Generate certificate file (`graylog.crt`):

```bash
openssl x509 -req -days 365 -in graylog.csr -signkey graylog.key -out graylog.crt -extensions v3_req -extfile graylog.cnf
```

Also, you must convert the private key to Graylog compatible PKCS#8 format (`graylog-8.key`):

```bash
openssl pkcs8 -in graylog.key -topk8 -nocrypt -out graylog-8.key
```

After these steps you can use `graylog.crt` as a Graylog certificate and `graylog-8.key` as a Graylog private key.
If you [configure TLS for Graylog HTTP interface into Cloud](#configure-tls-for-graylog-http-interface-into-cloud)
you should configure `.graylog.tls.http.cert` parameter for the certificate and `.graylog.tls.http.key` parameter for
the private key.

# Logging deploy schemes

This section describes base deployment schemes of Logging in Cloud and which connections can be closed by TLS.

## Graylog on VM

This deployment schema now positioned as default.

![TLS: Graylog on VM](/docs/images/tls/graylog-vm.png)

In this schema:

* Graylog and ElasticSearch deployed on separated VM as single docker containers
* Fluentd (or FluentBit) deployed in Kubernetes / OpenShift / another cloud as DaemonSet, collect logs and send them
  to Graylog outside the cloud.
* Other pods like cloud-events-reader also deployed in  Kubernetes / OpenShift / another cloud.

In this schema Logging support TLS for next channels:

| Connection                  | Status        |
| --------------------------- | ------------- |
| User - Graylog UI           | ✅ Support     |
| Graylog - ElasticSearch     | ✗ Not Support |
| Fluentd/FluentBit - Graylog | ✅ Support     |
| Fluentd/FluentBit - Logs    | N/A           |

where:

* `N/A` is `Not Applicable` because Fluentd/FluentBit read logs directly from containers log files on node.

Details how to configure TLS for this deployment schema see at sections:

* [Configure TLS for Graylog UI on VM](#configure-tls-for-graylog-ui-on-vm)
* [Configure TLS for Graylog inputs on VM](#configure-tls-for-graylog-inputs-on-vm)
* [Configure TLS for Fluentd output](#configure-tls-for-fluentd-output) or
  [Configure TLS for FluentBit output](#configure-tls-for-fluentbit-output)

## Graylog in Cloud

![TLS: Graylog in Cloud](/docs/images/tls/graylog-cloud.png)

In this schema:

* Graylog deployed in Kubernetes / OpenShift / another cloud as Deployment
* ElasticSearch **does not deploy** with Logging and should be use already deployed.
* Fluentd (or FluentBit) deployed in Kubernetes / OpenShift / another cloud as DaemonSet, collect logs and send them
  to Graylog outside the cloud.
* Other pods like cloud-events-reader also deployed in  Kubernetes / OpenShift / another cloud.

In this schema Logging support TLS for next channels:

| Connection                  | Status                |
| --------------------------- | --------------------- |
| User - Graylog UI           | ✅ Support             |
| Graylog - ElasticSearch     | ✅ Support             |
| Fluentd/FluentBit - Graylog | ✅/✗ Partially Support |
| Fluentd/FluentBit - Logs    | N/A                   |

where:

* `N/A` is `Not Applicable` because Fluentd/FluentBit read logs directly from containers log files on node.
* `✅/✗ Partially Support` means that for Graylog in Cloud currently there is no ability to configure TLS Input.
  But this feature in plans and will be implemented soon.

And one more important difference from schema when Graylog deploy on VM is TLS for Graylog UI. In the Cloud we have
Ingress which provide access outside the Cloud to applications UI into Cloud.

Ingress works as reverse proxy and can wrap traffic into TLS without actions from applications side. So for Graylog
in Cloud there is no make sense to configure TLS for UI. TLS will work through Ingress.

But we have to enable TLS for HTTPS interface in the Graylog configuration for ElasticSearch to secure the output
connection.

Details how to configure TLS for this deployment schema see at sections:

* [Configure TLS for Graylog Inputs into Cloud](#configure-tls-for-graylog-inputs-into-cloud)
* [Configure TLS for Graylog HTTP interface into Cloud](#configure-tls-for-graylog-http-interface-into-cloud)
  (e.g. for ElasticSearch connection)
* [Configure TLS for Fluentd output](#configure-tls-for-fluentd-output) or
  [Configure TLS for FluentBit output](#configure-tls-for-fluentbit-output)

## Configure TLS for Logging

This section describe how to configure TLS for various components into Logging stack.

### Configure TLS for Graylog UI on VM

TLS for Graylog UI works through Nginx reverse proxy.

To configure TLS for Graylog UI need execute next actions:

1. Prepare your certificates and package them into ZIP archive.
2. Upload early created ZIP archive into jos with using file choose dialog for parameter `certificates.zip`.
3. Specify the following parameters in `PARAMETERS_YAML`:

```ini
ssl_certificate=path/to/certificate/into/archive.crt
ssl_certificate_key=path/to/certificate/into/archive.pem
```

for example:

```ini
ssl_certificate=nginx/logging.crt
ssl_certificate_key=nginx/logging.pem
```

where:

* `ssl_certificate` is TLS certificate filename (.crt) that is presented on Deploy VM from the `certificates.zip` zip
  archive. Required for HTTPS UI access. If empty - self-signed certificate will be generated automatically.
* `ssl_certificate_key` is TLS certificate key filename (.pem) that is presented on Deploy VM from
  the `certificates.zip` zip archive. Required for HTTPS UI access.

### Configure TLS for Graylog Inputs on VM

TLS for Graylog Inputs works directly through Graylog. Third party applications are not used.

To configure TLS between Graylog and Fluentd/FluentBit need to configure TLS settings on both sides:

* Into Graylog, need to configure Graylog TLS Inputs
* Into Fluentd/FluentBit, need to configure TLS output for Graylog

To configure TLS for Graylog Input need execute the next actions:

1. Prepare your certificates and package them into ZIP archive.
2. Upload early created ZIP archive into jos with using file choose dialog for parameter `certificates.zip`.
3. Specify the following parameters in `PARAMETERS_YAML`:

```ini
tls_enabled=true
tls_cert_file=path/to/certificate/into/archive.crt
tls_key_file=path/to/key/into/archive.crt
tls_key_password=certificate_password
```

for example:

```ini
tls_enabled=true
tls_cert_file=graylog/logging.crt
tls_key_file=graylog/logging.pem
tls_key_password=awesome_password
```

* `tls_enabled` allow to enable TLS for Graylog input. The possible values are `true`/`false`.
* `tls_cert_file` is TLS certificate filename (.crt) that is presented on Deploy VM from the `certificates.zip`
  zip archive. Required for input with TLS.
* `tls_key_file` is TLS certificate key filename (.pem or .key) that is presented on Deploy VM from
  the `certificates.zip` zip archive. Required for input with TLS.
* `tls_key_password` the password for the TLS certificate key file for Graylog input.

**Note (not recommended, may be insecure)**: If you set `tls_enabled` to `true`, but _did not_ set
cert and key, Graylog will use self-signed certificate. Then you have to skip certificate verification
on Fluentd/FluentBit side.

If you want to enable/disable TLS for Graylog input quickly, you can do it through Graylog UI with the following steps:
login to Graylog UI -> `System/Inputs` -> `Inputs` -> choose the input and click on `More actions` -> `Edit input`
-> check or uncheck `Enable TLS`.

### Configure TLS for Graylog Inputs into Cloud

To configure TLS between Graylog and Fluentd/FluentBit need to configure TLS settings on both sides:

* Into Graylog, need to configure TLS for Graylog out-of-box GELF input
* Into Fluentd/FluentBit, need to configure TLS for out-of-box GELF output for Graylog

To configure TLS for Graylog Inputs need to execute the following actions:

1. Create PEM encoded certificates for Graylog. It's better to use certificates signed by any CA, not self-signed.
2. Add early created certificate and private key into Secret and create this secret. For example:

    ```yaml
    kind: Secret
    apiVersion: v1
    type: Opaque
    metadata:
      name: graylog-input-tls
    data:
      input.crt: ...base64 encoded PEM certificate...
      input.key: ...base64 encoded PEM key...
    ```

3. During deploy need specify the following deploy parameters:

    ```yaml
    graylog:
      tls:
        input:
          enabled: true
          cert:
            secretName: graylog-input-tls
            secretKey: input.crt
          key:
            secretName: graylog-input-tls
            secretKey: input.key
          keyFilePassword: password # Optional. The password to unlock the private key (input.key in graylog-input-tls)
    ```

4. Configure [Fluentd](#configure-tls-for-fluentd-output) of [FluentBit](#configure-tls-for-fluentbit-output)
   outputs. You have to use CA certificate for your Graylog Input certificate as CA certificate during
   Fluentd/FluentBit installation. Otherwise, you can skip verification by using the corresponding parameter
   **(not recommended, may be insecure)**.

**Note (not recommended, may be insecure)**: If you set `.graylog.tls.input.enabled` to `true`, but _did not_ set
cert and key, Graylog will use self-signed certificate. Then you have to skip certificate verification
on Fluentd/FluentBit side.

If you want to enable/disable TLS for Graylog input quickly, you can do it through Graylog UI with the following steps:
login to Graylog UI -> `System/Inputs` -> `Inputs` -> choose the input and click on `More actions` -> `Edit input`
-> check or uncheck `Enable TLS`.

Parameters that configure TLS for Graylog Inputs stored in the `.graylog.tls.input` section:

* `.graylog.tls.input.enabled` - Allows enabling TLS for out-of-box GELF input managed by the
  operator. The data type is `Bool`. The default value is `false`.
* `.graylog.tls.input.generateCerts` - Allows enabling integration with [cert-manager](#about-cert-manager).
  The data type is `object`.
* `.graylog.tls.input.cert.secretName` - Name of Secret with the certificate. The data type is `Str`.
* `.graylog.tls.input.cert.secretKey` - Key (filename) in the Secret with the certificate. The data type is `Str`.
* `.graylog.tls.input.key.secretName` - Name of Secret with the private key for the certificate. The data type is `Str`.
* `.graylog.tls.input.key.secretKey` - Key (filename) in the Secret with the private key for the certificate.
  The data type is `Str`.
* `.graylog.tls.input.keyFilePassword` - The password to unlock the private key used for securing Graylog input.
  The data type is `Str`.

### Configure TLS for Graylog HTTP interface into Cloud

To configure TLS for HTTP interface (it's used for ElasticSearch) need to execute the following actions:

1. Create certificates for Graylog. Note that Graylog works only with **PEM** encoded certificates and **PEM** encoded
   private keys in **PKCS#8** format. Look at the [official Graylog documentation](https://docs.graylog.org/docs/https)
   for additional information. It's better to use certificates signed by any CA, not self-signed.
   **Warning**: Graylog requires the use of a certificate that can verify the
   `graylog-service.<your_k8s_namespace>.svc` hostname. It means that if you deploy a logging-operator to the
   `logging` namespace on Kubernetes, you should add the `graylog-service.logging.svc` hostname as one of the alt_names
   during creating the certificate. Otherwise, you may see the following messages in Graylog logs:

   ```bash
   Hostname graylog-service.logging.svc not verified...
   ```

2. Add early created certificate and private key in PKCS#8 format into Secret and create this secret. For example:

    ```yaml
    kind: Secret
    apiVersion: v1
    type: Opaque
    metadata:
      name: graylog-http-tls
    data:
      graylog.crt: ...base64 encoded PEM certificate...
      graylog.key: ...base64 encoded PEM key in PKCS#8 format...
    ```

3. (**optional**) If external services that connect to your Graylog (e.g. ElasticSearch) use certificates, you can add
   CA certificates for it to the Graylog's keystore. You need to create Secret and put all CA certificates to separate
   keys in it. For example:

    ```yaml
    kind: Secret
    apiVersion: v1
    type: Opaque
    metadata:
      name: graylog-ca-certs
    data:
      ca1.crt: ...base64 encoded certificate 1...
      ca2.crt: ...base64 encoded certificate 2...
      ...
    ```

   Then you should set the name of Secret to the `.graylog.tls.http.cacerts` parameter.

   **Note**: Certificate that you use for Graylog will be added to the keystore automatically.

4. During deploy need specify the following deploy parameters:

    ```yaml
    graylog:
      tls:
        http:
          enabled: true
          cacerts: graylog-ca-certs # Optional
          cert:
            secretName: graylog-http-tls
            secretKey: graylog.crt
          key:
            secretName: graylog-http-tls
            secretKey: graylog.key
          keyFilePassword: password # Optional. The password to unlock the private key (graylog.key in graylog-http-tls)
    ```

**Note**: If you set `.graylog.tls.http.enabled` to `true`, you **have to** set cert and key for correct work.

Parameters that configure TLS for Graylog HTTP interface stored in the `.graylog.tls.http` section:

* `.graylog.tls.http.enabled` - Allows enabling TLS for HTTP interface. If this parameter is true,
  each connection to and from the Graylog server except inputs will be secured by TLS, including API calls of the server
  to itself. The data type is `Bool`.The default value is `false`.
* `.graylog.tls.http.generateCerts` - Allows enabling integration with [cert-manager](#about-cert-manager).
  The data type is `object`.
* `.graylog.tls.http.cacerts` - Contains a name of Secret with CA certificates for a custom CA store.
  If present, all certificates from the Secret will be added to the Java keystore.
  The keystore can be used for TLS in the custom inputs as well. The data type is `Str`.
* `.graylog.tls.http.cert.secretName` - (**Required, if `.graylog.tls.http.enabled` is true**)
  Name of Secret with the certificate. This certificate will be added to the Java keystore. The data type is `Str`.
* `.graylog.tls.http.cert.secretKey` - (**Required, if `.graylog.tls.http.enabled` is true**)
  Key (filename) in the Secret with the certificate. The data type is `Str`.
* `.graylog.tls.http.key.secretName` - (**Required, if `.graylog.tls.http.enabled` is true**)
  Name of Secret with the private key for the certificate. The data type is `Str`.
* `.graylog.tls.http.key.secretKey` - (**Required, if `.graylog.tls.http.enabled` is true**)
  Key (filename) in the Secret with the private key for the certificate. The data type is `Str`.
* `.graylog.tls.http.keyFilePassword` - The password to unlock the private key used for securing the HTTP interface.
  The data type is `Str`.

### Configure TLS for Fluentd output

To configure TLS for Fluentd output need to execute the following actions:

1. Create certificates for Fluentd. It's better to use certificates signed by any CA, not self-signed.
2. Add early created certificates into Secret and create this secret. For example:

    ```yaml
    kind: Secret
    apiVersion: v1
    type: Opaque
    metadata:
      name: fluentd-tls-certificates
    data:
      fluentd-ca.crt: ...base64 encoded certificate...
      fluentd.crt: ...base64 encoded certificate...
      fluentd.key: ...base64 encoded certificate...
    ```

3. During deploy need specify the following deploy parameters:

    ```yaml
    fluentd:
      tls:
        enabled: true
        noDefaultCA: false
        version: :TLSv1_2
        ca:
          secretName: fluentd-tls-certificates
          secretKey: fluentd-ca.crt
        cert:
          secretName: fluentd-tls-certificates
          secretKey: fluentd.crt
        key:
          secretName: fluentd-tls-certificates
          secretKey: fluentd.key
        allCiphers: true
        rescueSslErrors: false
        noVerify: false
    ```

**Note**: Parameter which default values you doesn't want to change or can skip and doesn't specify
in deploy parameters.

where:

* `.fluentd.tls.enabled` - Allows enabling TLS for out-of-box Graylog GELF output managed by
  the operator. The data type is `Bool`. The default value is `false`.
* `.fluentd.tls.generateCerts` - Allows enabling integration with [cert-manager](#about-cert-manager).
  The data type is `object`.
* `.fluentd.tls.noDefaultCA` - Prevents OpenSSL from using the systems CA store. The data type is `Bool`.
  The default value is `false`.
* `.fluentd.tls.version` - Any of :TLSv1, :TLSv1_1, :TLSv1_2. The data type is `Str`. The default value is `:TLSv1_2`.
* `.fluentd.tls.ca.secretName` - Name of Secret with CA certificate. The data type is `Str`.
* `.fluentd.tls.ca.secretKey` - Key (filename) in the Secret with CA certificate. The data type is `Str`.
* `.fluentd.tls.cert.secretName` - Name of Secret with client certificate. The data type is `Str`.
* `.fluentd.tls.cert.secretKey` - Key (filename) in the Secret with client certificate. The data type is `Str`.
* `.fluentd.tls.key.secretName` - Name of Secret with key for the client certificate. The data type is `Str`.
* `.fluentd.tls.key.secretKey` - Key (filename) in the Secret with key for the client certificate.
  The data type is `Str`.
* `.fluentd.tls.allCiphers` - Allows any ciphers to be used, may be insecure. The data type is `Bool`.
  The default value is `true`.
* `.fluentd.tls.rescueSslErrors` - Similar to rescue_network_errors in notifier.rb, allows SSL exceptions to be raised.
  The data type is `Bool`. The default value is `false`.
* `.fluentd.tls.noVerify` - Disable peer verification. The data type is `Bool`. The default value is `false`.

### Configure TLS for FluentBit output

FluentBit has another parameters to enable TLS output:

1. Create certificates for Fluentd. It's better to use certificates signed by any CA, not self-signed.
2. Add early created certificates into Secret and create this secret. For example:

    ```yaml
    kind: Secret
    apiVersion: v1
    type: Opaque
    metadata:
      name: fluentbit-tls-certificates
    data:
      fluentbit-ca.crt: ...base64 encoded certificate...
      fluentbit.crt: ...base64 encoded certificate...
      fluentbit.key: ...base64 encoded certificate...
    ```

3. During deploy need specify the following deploy parameters:

    ```yaml
    fluentbit:
      tls:
        enabled: true
        ca:
          secretName: fluentbit-tls-certificates
          secretKey: fluentbit-ca.crt
        cert:
          secretName: fluentbit-tls-certificates
          secretKey: fluentbit.crt
        key:
          secretName: fluentbit-tls-certificates
          secretKey: fluentbit.key
        verify: true
        keyPasswd: <optional_password>
    ```

where:

* `.fluentbit.tls.enabled` - Allows enabling TLS for out-of-box Graylog GELF output managed by
  the operator. The data type is `Bool`. The default value is `false`.
* `.fluentbit.tls.generateCerts` - Allows enabling integration with [cert-manager](#about-cert-manager).
  The data type is `object`.
* `.fluentbit.tls.ca.secretName` - Name of Secret with CA certificate. The data type is `Str`.
* `.fluentbit.tls.ca.secretKey` - Key (filename) in the Secret with CA certificate. The data type is `Str`.
* `.fluentbit.tls.cert.secretName` - Name of Secret with client certificate. The data type is `Str`.
* `.fluentbit.tls.cert.secretKey` - Key (filename) in the Secret with client certificate. The data type is `Str`.
* `.fluentbit.tls.key.secretName` - Name of Secret with key for the client certificate. The data type is `Str`.
* `.fluentbit.tls.key.secretKey` - Key (filename) in the Secret with key for the client certificate.
  The data type is `Str`.
* `.fluentbit.tls.verify` - Force certificate validation. The data type is `Bool`. The default value is `true`.
* `.fluentbit.tls.keyPasswd` - Optional password for private key file. The data type is `Str`.

### About cert-manager

[Cert-manager](https://cert-manager.io/) is X.509 certificate controller for Kubernetes and OpenShift workloads. It will
obtain certificates from a variety of Issuers, both popular public Issuers as well as private Issuers, and ensure
the certificates are valid and up-to-date, and will attempt to renew certificates at a configured time before expiry.

In order for cert-manager to generate a secret containing certificates and private key, you need to take several steps:

0. Make sure cert-manager is installed on the cluster. It is usually installed in the `cert-manager` namespace.
1. You should use [Issuer or ClusterIssuer](https://cert-manager.io/docs/concepts/issuer/) resource for creating
   certificate. These resources represent certificate authorities (CAs) that are able to generate signed certificates
   by honoring certificate signing requests. Example of namespaced self-signed Issuer resource:

   ```yaml
   apiVersion: cert-manager.io/v1
   kind: Issuer
   metadata:
     name: logging-self-signed-issuer
   spec:
     selfSigned: {}
   ```

2. Then you can create [Certificate](https://cert-manager.io/docs/concepts/certificate/) resource. Configuration of this
   resource allows to change parameters of generated certificates and private key. You can find an example of
   certificate resource [here](https://cert-manager.io/docs/usage/certificate/#creating-certificate-resources).
3. Cert-manager will create [Certificate Request](https://cert-manager.io/docs/concepts/certificaterequest/)
   resource based on created Certificate resource.
4. Also, cert-manager will create Secret resource with name specified in the Certificate resource previously.
   Generated secret contains fields `ca.crt` with PEM CA certificate, `tls.crt` with PEM private key and `tls.key`
   with PEM signed certificate chain by default.

The generated secret can be used in pods later as a volume.

#### Logging-operator integration with cert-manager

If you use logging-operator, you don't need to create cert-manager resources for its components manually.

Each component of logging-operator that has TLS configuration can use cert-manager and has the same configuration for
it. You enable and customise work with cert-manager separately for Fluentd, Fluent-bit, Graylog HTTP server or inputs
by the same in a similar way via the `tls.generateCerts` section.

Integration with cert-manager can be configured with `fluentd.tls.generateCerts`, `fluentbit.tls.generateCerts`,
`graylog.tls.http.generateCerts` and `graylog.tls.input.generateCerts`. The only required parameter to use cert-manager
is `generateCerts.enabled`, which must be set to true. Each `generateCerts` section has the following parameters:

* `generateCerts.enabled`: enables or disables integration of component with cert-manager.
* `generateCerts.secretName`: defines the name of the generated Secret (default values: `fluentd-cert-manager-tls` for
  Fluentd, `fluentbit-cert-manager-tls` for Fluent-bit, `graylog-http-cert-manager-tls` for Graylog HTTP,
  `graylog-input-cert-manager-tls` for Graylog inputs).
* `generateCerts.clusterIssuerName`: allows to specify pre-created ClusterIssuer instead of Issuer created by Helm to
  verify generated certs. Notice that ClusterIssuer is a cluster-wide entity and should be created manually before
  deploy in any namespace on cluster. If `clusterIssuerName` is empty (by default), Helm will create Issuer with
  single self-signed CA certificate for all components. See more info below.
* `generateCerts.duration`: allows configuring duration of generated certificates
  (integer value in **days**; `365` by default).
* `generateCerts.renewBefore`: specifies how long before expiry a certificate should be renewed
  (integer value in **days**; `15` by default).

**NOTE**: If you want to connect Fluentd or Fluent-bit and Graylog with TLS, you must ensure that the certificates pass
verification during the handshake. If you're using cert-manager, you can use the same Issuer/ClusterIssuer for all
components.

If generateCerts is enabled for at least one component, Helm will create an Issuer with a self-signed CA certificate
inside the namespace containing the logging-operator during deploy. In more detail, the following resources will be
created:

1. Issuer `logging-self-signed-issuer` that can be used for self-signed certs creation;
2. CA certificate `logging-self-signed-ca-certificate` based on `logging-self-signed-issuer` Issuer;
3. Secret `logging-self-signed-ca-certificate` that contains created CA certificate;
4. Issuer `logging-ca-issuer` based on `logging-self-signed-ca-certificate` CA certificate.

Issuer `logging-ca-issuer` will be used for every certificate generated by cert-manager for logging-operator, if
`clusterIssuerName` is not specified for the component.

# Examples

## Graylog in cloud with manually created secrets

```yaml
graylog:
  tls:
    input:
      enabled: true
      cert:
        secretName: graylog-input-tls
        secretKey: input.crt
      key:
        secretName: graylog-input-tls
        secretKey: input.key
      keyFilePassword: password # Optional
    http:
      enabled: true
      cacerts: graylog-ca-certs # Optional
      cert:
        secretName: graylog-http-tls
        secretKey: graylog.crt
      key:
        secretName: graylog-http-tls
        secretKey: graylog.key
      keyFilePassword: password # Optional
```

## Graylog in cloud with cert-manager integration

```yaml
graylog:
  tls:
    input:
      enabled: true
      generateCerts:
        enabled: true
        secretName: graylog-input-cert-manager-tls # Optional
        clusterIssuerName: ""                      # Optional
        duration: 365                              # Optional
        renewBefore: 15                            # Optional
    http:
      enabled: true
      generateCerts:
        enabled: true
        secretName: graylog-http-cert-manager-tls # Optional
        clusterIssuerName: ""                     # Optional
        duration: 365                             # Optional
        renewBefore: 15                           # Optional
```

## Fluentd in cloud with manually created secrets

```yaml
fluentd:
  tls:
    enabled: true
    noDefaultCA: false    # Optional
    version: :TLSv1_2     # Optional
    ca:
      secretName: fluentd-tls-certificates
      secretKey: fluentd-ca.crt
    cert:
      secretName: fluentd-tls-certificates
      secretKey: fluentd.crt
    key:
      secretName: fluentd-tls-certificates
      secretKey: fluentd.key
    allCiphers: true       # Optional
    rescueSslErrors: false # Optional
    noVerify: false        # Optional
```

## Fluentd in cloud with cert-manager integration

```yaml
fluentd:
  tls:
    enabled: true
    generateCerts:
      enabled: true
      secretName: fluentd-cert-manager-tls # Optional
      clusterIssuerName: ""                # Optional
      duration: 365                        # Optional
      renewBefore: 15                      # Optional
```

## Fluent-bit in cloud with manually created secrets

```yaml
fluentbit:
  tls:
    enabled: true
    ca:
      secretName: fluentbit-tls-certificates
      secretKey: fluentbit-ca.crt
    cert:
      secretName: fluentbit-tls-certificates
      secretKey: fluentbit.crt
    key:
      secretName: fluentbit-tls-certificates
      secretKey: fluentbit.key
    verify: true        # Optional
    keyPasswd: password # Optional
```

## Fluent-bit in cloud with cert-manager integration

```yaml
fluentbit:
  tls:
    enabled: true
    generateCerts:
      enabled: true
      secretName: fluentbit-cert-manager-tls # Optional
      clusterIssuerName: ""                  # Optional
      duration: 365                          # Optional
      renewBefore: 15                        # Optional
```
