This document provides information about deploy Logging Service in AWS Kubernetes Service and configure AWS
OpenSearch as storage.

# Table of Content

* [Table of Content](#table-of-content)
* [Before you begin](#before-you-begin)
* [Configuration](#configuration)
  * [Configure AWS OpenSearch](#configure-aws-opensearch)
    * [Fine-grained access control](#fine-grained-access-control)
      * [Use internal user database](#use-internal-user-database)
      * [IAM roles](#iam-roles)
  * [Deploy Logging Service](#deploy-logging-service)

# Before you begin

* [Architectural Considerations](https://go2docs.graylog.org/4-x/planning_your_deployment/planning_your_deployment.html)
* Graylog >= 4.3 required. We include it in Logging since:
  * Either `Logging Application` >= **v.10.9.0**
  * Or `logging-operator` >= **13.4.0**
* OpenSearch 2.x **isn't officially supported** by Graylog 4.x. See offical documenation
  [Upgradiing to OpenSearch](https://go2docs.graylog.org/4-x/planning_your_deployment/upgrading_to_opensearch_-_installation.htm?Highlight=Opensearch)
* Recommended use OpenSearch with resources not less that:
  * CPU: >= 2 cores
  * Memory: >= 8 Gb

# Configuration

This document describe how to configure AWS and Logging to deploy.

## Configure AWS OpenSearch

Before you start deploy Logging need to create AWS OpenSearch domain. Or you also can use already
exists domain.

To check that domain already exists or create new via AWS Console need:

1. Go to [https://aws.amazon.com](https://aws.amazon.com/) and choose Sign In to the Console.
2. Under Analytics, choose Amazon OpenSearch Service.
3. Choose Create domain.
4. For Domain name, enter a domain name. The name must meet the following criteria:
    * Unique to your account and AWS Region
    * Starts with a lowercase letter
    * Contains between 3 and 28 characters
    * Contains only lowercase letters a-z, the numbers 0-9, and the hyphen (-)
5. If you want to use a custom endpoint rather than the standard one of `https://search-mydomain-1a2a3a4a5a6a7a8a9a0a9a8a7a.us-east-1.es.amazonaws.com`,
   choose Enable custom endpoint and provide a name and certificate. For more information,
   see [Creating a custom endpoint for Amazon OpenSearch Service](https://docs.aws.amazon.com/opensearch-service/latest/developerguide/customendpoint.html).
6. For Deployment type, choose the option that best matches the purpose of your domain:
    * Production domains use Multi-AZ and dedicated master nodes for higher availability.
    * Development and testing domains use a single Availability Zone.
    * Custom domains let you choose from all configuration options.

   **Important**: Different deployment types present different options on subsequent pages.
   These steps include all options (the Custom deployment type).

7. For **Version**, choose the version of OpenSearch or legacy Elasticsearch OSS to use.
   We recommend that you choose the latest version of OpenSearch. For more information,
   see [Supported versions of OpenSearch and Elasticsearch](https://docs.aws.amazon.com/opensearch-service/latest/developerguide/what-is.html#choosing-version).

   (Optional) If you chose an OpenSearch version for your domain, select **Enable compatibility mode**
   to make OpenSearch report its version as 7.10, which allows certain Elasticsearch OSS clients
   and plugins that check the version before connecting to continue working with the service.

8. For **Auto-Tune**, choose whether to allow OpenSearch Service to suggest memory-related
   configuration changes to your domain to improve speed and stability. For more information,
   see [Auto-Tune for Amazon OpenSearch Service](https://docs.aws.amazon.com/opensearch-service/latest/developerguide/auto-tune.html).

   (Optional) Select **Add maintenance window** to schedule a recurring window during which
   which Auto-Tune updates the domain.

9. Under **Data nodes**, choose the number of availability zones. For more information, see
   [Configuring a multi-AZ domain in Amazon OpenSearch Service](https://docs.aws.amazon.com/opensearch-service/latest/developerguide/managedomains-multiaz.html).

   **Note**: The OpenSearch Service console doesn't support moving from multiple availability zones
   to a single availability zone after the domain is created. If you choose 2 or 3 availability zones
   and later want to move to 1, you must disable the `ZoneAwarenessEnabled` parameter using
   the AWS CLI or configuration API.

10. For **Instance type**, choose an instance type for your data nodes. For more information, see
    [Supported instance types in Amazon OpenSearch Service](https://docs.aws.amazon.com/opensearch-service/latest/developerguide/supported-instance-types.html).

    **Note**: Not all Availability Zones support all instance types. If you choose **3-AZ**,
    we recommend choosing current-generation instance types such as R5 or I3.

11. For **Number** of nodes, choose the number of data nodes.

    For maximum values, see [Domain and instance quotas](https://docs.aws.amazon.com/opensearch-service/latest/developerguide/limits.html#clusterresource).
    Single-node clusters are fine for development and testing, but should not be used for production
    workloads. For more guidance, see [Sizing Amazon OpenSearch Service domains](https://docs.aws.amazon.com/opensearch-service/latest/developerguide/sizing-domains.html)
    and [Configuring a multi-AZ domain in Amazon OpenSearch Service](https://docs.aws.amazon.com/opensearch-service/latest/developerguide/managedomains-multiaz.html).

12. For **Storage type**, select Amazon EBS or instance store volumes to associate with your instance.
    The volume types available in the list depend on the instance type that you've chosen.
    For guidance on creating especially large domains, see
    [Petabyte scale in Amazon OpenSearch Service](https://docs.aws.amazon.com/opensearch-service/latest/developerguide/petabyte-scale.html).

13. If you chose EBS as the storage type, configure the following additional settings.
    Some settings might not appear depending on the type of volume you choose.

    <!-- markdownlint-disable line-length -->
    | Setting                       | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     |
    | ----------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
    | **EBS volume type**           | Choose between [General Purpose (SSD) - gp3](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ebs-volume-types.html#gp3-ebs-volume-type) and [General Purpose (SSD) - gp2](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ebs-volume-types.html#EBSVolumeTypes_gp2), or the previous generation [Provisioned IOPS (SSD)](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ebs-volume-types.html#EBSVolumeTypes_piops), and [Magnetic (standard)](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ebs-volume-types.html#EBSVolumeTypes_standard). |
    | **EBS storage size per node** | Enter the size of the EBS volume that you want to attach to each data node. EBS volume size is per node. You can calculate the total cluster size for the OpenSearch Service domain by multiplying the number of data nodes by the EBS volume size. The minimum and maximum size of an EBS volume depends on both the specified EBS volume type and the instance type that it's attached to. To learn more, see [EBS volume size limits](https://docs.aws.amazon.com/opensearch-service/latest/developerguide/limits.html#ebsresource).                         |
    | **Provisioned IOPS**          | If you selected a Provisioned IOPS SSD volume type, enter the number of I/O operations per second (IOPS) that the volume can support.                                                                                                                                                                                                                                                                                                                                                                                                                           |
    <!-- markdownlint-enable line-length -->

14. (Optional) If you selected a gp3 volume type, expand Advanced settings and specify additional
    IOPS (up to 1,000 MiB/s) and throughput (up to 16,000) to provision for each node, beyond
    what is included with the price of storage, for an additional cost. For more information, see the
    [Amazon OpenSearch Service pricing](https://aws.amazon.com/opensearch-service/pricing/).
15. Choose the type and number of [dedicated master nodes](https://docs.aws.amazon.com/opensearch-service/latest/developerguide/managedomains-dedicatedmasternodes.html).
    Dedicated master nodes increase cluster stability and are required for domains that have instance
    counts greater than 10. We recommend three dedicated master nodes for production domains.

    **Note**: You can choose different instance types for your dedicated master nodes and data nodes.
    For example, you might select general purpose or storage-optimized instances for your data nodes,
    but compute-optimized instances for your dedicated master nodes.

16. (Optional) To enable [UltraWarm storage](https://docs.aws.amazon.com/opensearch-service/latest/developerguide/ultrawarm.html),
    choose **Enable UltraWarm data nodes**. Each instance type has a [maximum amount of storage](https://docs.aws.amazon.com/opensearch-service/latest/developerguide/limits.html#limits-ultrawarm)
    that it can address. Multiply that amount by the number of warm data nodes for the total
    addressable warm storage.

17. (Optional) To enable [cold storage](https://docs.aws.amazon.com/opensearch-service/latest/developerguide/cold-storage.html),
    choose **Enable cold storage**. You must enable UltraWarm to enable cold storage.

18. (Optional) For domains running OpenSearch or Elasticsearch 5.3 and later, the **Snapshot configuration**
    is irrelevant. For more information about automated snapshots, see
    [Creating index snapshots in Amazon OpenSearch Service](https://docs.aws.amazon.com/opensearch-service/latest/developerguide/managedomains-snapshots.html).

19. Under **Network**, choose either **VPC access** or **Public access**. If you choose **Public access**,
    skip to the next step. If you choose **VPC access**, make sure you meet the prerequisites,
    then configure the following settings:

    <!-- markdownlint-disable line-length -->
    | Setting             | Description                                                                                                                                                                                                                                                                                                                                                                                                                                     |
    | ------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
    | **VPC**             | Choose the ID of the virtual private cloud (VPC) that you want to use. The VPC and domain must be in the same AWS Region, and you must select a VPC with tenancy set to **Default**. OpenSearch Service does not yet support VPCs that use dedicated tenancy.                                                                                                                                                                                   |
    | **Subnet**          | Choose a subnet. If you enabled Multi-AZ, you must choose two or three subnets. OpenSearch Service will place a VPC endpoint and elastic network interfaces in the subnets. You must reserve sufficient IP addresses for the network interfaces in the subnet(s). For more information, see [Reserving IP addresses in a VPC subnet](https://docs.aws.amazon.com/opensearch-service/latest/developerguide/vpc.html#reserving-ip-vpc-endpoints). |
    | **Security groups** | Choose one or more VPC security groups that allow your required application to reach the OpenSearch Service domain on the ports (80 or 443) and protocols (HTTP or HTTPs) exposed by the domain. For more information, see [Launching your Amazon OpenSearch Service domains within a VPC](https://docs.aws.amazon.com/opensearch-service/latest/developerguide/vpc.html).                                                                      |
    | **IAM Role**        | Keep the default role. OpenSearch Service uses this predefined role (also known as a service-linked role) to access your VPC and to place a VPC endpoint and network interfaces in the subnet of the VPC. For more information, see [Service-linked role for VPC access](https://docs.aws.amazon.com/opensearch-service/latest/developerguide/vpc.html#enabling-slr).                                                                           |
    <!-- markdownlint-enable line-length -->

20. Enable or disable fine-grained access control:
    * If you want to use IAM for user management, choose **Set IAM ARN as master user** and specify
      the ARN for an IAM role.
    * If you want to use the internal user database, choose **Create master user** and specify a user
      name and password.

    Whichever option you choose, the master user can access all indexes in the cluster and all OpenSearch APIs.
    For guidance on which option to choose, see [Key concepts](https://docs.aws.amazon.com/opensearch-service/latest/developerguide/fgac.html#fgac-concepts).
    If you disable fine-grained access control, you can still control access to your domain by placing it
    within a VPC, applying a restrictive access policy, or both. You must enable node-to-node encryption
    and encryption at rest to use fine-grained access control.

    **Note**: We *strongly* recommend enabling fine-grained access control to protect the data on your domain.
    Fine-grained access control provides security at the cluster, index, document, and field levels.

21. (Optional) If you want to use SAML authentication for OpenSearch Dashboards, choose **Prepare SAML authentication**.
    After the domain is available, see [SAML authentication for OpenSearch Dashboards](https://docs.aws.amazon.com/opensearch-service/latest/developerguide/saml.html)
    for additional steps.

22. (Optional) If you want to use Amazon Cognito authentication for OpenSearch Dashboards, choose
    **Enable Amazon Cognito authentication**. Then choose the Amazon Cognito user pool and identity pool
    that you want to use for OpenSearch Dashboards authentication. For guidance on creating these resources, see
    [Configuring Amazon Cognito authentication for OpenSearch Dashboards](https://docs.aws.amazon.com/opensearch-service/latest/developerguide/cognito-auth.html).

23. For **Domain access policy**, choose an access policy or configure one of your own. If you choose
    to create a custom policy, you can configure it yourself or import one from another domain.
    For more information, see [Identity and Access Management in Amazon OpenSearch Service](https://docs.aws.amazon.com/opensearch-service/latest/developerguide/ac.html).

    **Note**: If you enabled VPC access, you can't use IP-based policies. Instead, you can use [security groups](https://docs.aws.amazon.com/vpc/latest/userguide/VPC_SecurityGroups.html)
    to control which IP addresses can access the domain. For more information, see
    [About access policies on VPC domains](https://docs.aws.amazon.com/opensearch-service/latest/developerguide/vpc.html#vpc-security).

24. (Optional) To require that all requests to the domain arrive over HTTPS, select
    **Require HTTPS for all traffic to the domain**.

25. (Optional) To enable node-to-node encryption, select **Node-to-node encryption**. For more information, see
    [Node-to-node encryption for Amazon OpenSearch Service](https://docs.aws.amazon.com/opensearch-service/latest/developerguide/ntn.html).

26. (Optional) To enable encryption of data at rest, select **Enable encryption of data at rest**.
    Select **Use AWS owned key** to have OpenSearch Service create an AWS KMS encryption key on your behalf
    (or use the one that it already created). Otherwise, choose your own KMS key. For more information, see
    [Encryption of data at rest for Amazon OpenSearch Service](https://docs.aws.amazon.com/opensearch-service/latest/developerguide/encryption-at-rest.html).

27. (Optional) Add tags to describe your domain so you can categorize and filter on that information.
    For more information, see [Tagging Amazon OpenSearch Service domains](https://docs.aws.amazon.com/opensearch-service/latest/developerguide/managedomains-awsresourcetagging.html).

28. (Optional) Expand and configure **Advanced cluster settings**. For a summary of these options, see
    [Advanced cluster settings](https://docs.aws.amazon.com/opensearch-service/latest/developerguide/createupdatedomains.html#createdomain-configure-advanced-options).

29. Choose Create.

If you want to create or update Domain via AWS CLI or want to read original instruction, you can read
the official documentation [Creating and managing Amazon OpenSearch Service domains](https://docs.aws.amazon.com/opensearch-service/latest/developerguide/createupdatedomains.html).

### Fine-grained access control

For OpenSearch there some options
[Fine-grained access control](https://docs.aws.amazon.com/opensearch-service/latest/developerguide/fgac.html).

We can configure AWS OpenSearch domain to work with:

* Internal user database - when we configure admin user in AWS UI, and create other users into OpenSearch
  with using REST requests
* Identity-based policies - OpenSearch will use IAM and IAM Users or Roles for auth (*not supported case for now*)
* No authentication

#### Use internal user database

To configure domain use internal user database need to navigate:

1. Open AWS OpenSearch
2. Navigate Domains -> `<domain_name>`
3. Open tab "Security configuration"
4. Click "Edit" button
5. In section "Fine-grained access control" select "Create master user" and fill username and password
6. Click "Save" button

After these steps for login in AWS OpenSearch Dashboard you can use created admin user. And you can use
OpenSearch API to create new users.

For example you can use the following parameters:

```yaml
graylog:
  elasticsearchHost: https://<login>:<password>@<opensearch_url>
```

#### IAM roles

**Warning!** Currently Graylog doesn't allow use AWS OpenSearch with AWS IAM roles! Use other authentification options.

To configure domain use IAM roles need to navigate:

1. Open AWS OpenSearch
2. Navigate Domains -> `<domain_name>`
3. Open tab "Security configuration"
4. Click "Edit" button
5. In section "Fine-grained access control" select "Set IAM ARN as master user" and fill ARN for user
6. Click "Save" button

OpenSearch will use IAM and IAM Users or Roles for auth
but graylog has no native support of IAM Roles.
So we can't use IAM Users and IAM Roles with Logging Service.

## Deploy Logging Service

To deploy Logging in AWS Kubernetes Service and with using AWS OpenSearch you need prepare AWS OpenSearch
(see above). And next deploy service with the following parameters:

**Note**: Full list of parameters and descriptions for them you can read in installation guide
[Installation Guide](../installation.md).

```yaml
name: logging-service
createClusterAdminEntities: "true"
containerRuntimeType: containerd
graylog:
  install: true
  password: <password>
  host: https://<graylog-url>/
  inputPort: 12201
  elasticsearchHost: https://<login>:<password>@<vpc opensearch url>/
  graylogResources:
    requests:
      cpu: "400m"
      memory: "1500Mi"
    limits:
      cpu: "800m"
      memory: "2000Mi"
  mongoResources:
    requests:
      cpu: "300m"
      memory: "256Mi"
    limits:
      cpu: "500m"
      memory: "256Mi"
  mongoStorageClassName: <StorageClass>
  graylogStorageClassName: <StorageClass>
  storageSize: 10Gi
  contentDeployPolicy: only-create
  logsRotationSizeGb: 20
  javaOpts: -Xms1024m -Xmx1024m
fluentd:
  install: true
  systemLogType: "varlogsyslog"
  graylogHost: "<graylog service name>.<namespace>.svc"
  securityResources:
    name: logging-fluentd
    install: true
cloudEventsReader:
  install: true
```
