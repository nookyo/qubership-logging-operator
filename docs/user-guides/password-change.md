The document provides information about changing users\` passwords process in Graylog:

# Table of Content

* [Table of Content](#table-of-content)
* [Overview](#overview)
* [Change user password in Cloud](#change-user-password-in-cloud)
  * [Change password in Graylog UI](#change-password-in-graylog-ui)
  * [Change password using Graylog REST API](#change-password-using-graylog-rest-api)
* [Change root user password in Cloud](#change-root-user-password-in-cloud)
  * [Change password in Kubernetes Secret](#change-password-in-kubernetes-secret)

# Overview

Graylog has default root user usually named admin. Other users can be created and configured in UI. They have special
roles with pre-configured permissions.

# Change user password in Cloud

There are two ways of changing password:

* Using UI interface
* Using Graylog REST API

## Change password in Graylog UI

To change users password need to:

1. Login in Graylog UI (you have use a user that has permission to control other users)
2. Navigate to `System -> Users and Terms`
3. Find necessary user and open it profile
4. Click on the button `Edit User`
5. Find on the page `New Password` and `Repeat Password` fields and fill them
6. Click the button `Change Password`

## Change password using Graylog REST API

To change user password need to know it's ID that consists of 24 symbols (`[a-z0-9]`).
It can be found in url in browser on user's page or if you have grants you can get it by GET request
`http://<host>/api/users`.

To send request you can be authorized as user you want to change.

Request:

* PUT `https://<host>/api/users/<user_id>/password`
* Headers:

    ```json
    Content-Type: application/json
    X-Requested-By: Graylog API Browser
    ```

* Json body:

    ```json
    {
        "password": "new_password",
        "old_password": "old_password"
    }
    ```

Successful response code is `204`. It means that the password was successfully updated and user must make all next
actions with the new password.

According to Graylog API Browser there are possible error status codes:

* `400` - The new password is missing, or the old password is missing or incorrect.
* `403` - The requesting user has insufficient privileges to update the password for the given user.
* `404` - User does not exist.

# Change root user password in Cloud

The root password in Graylog can't be changes as credentials for regular users. The credentials for this user
stored in Graylog configuration file.

So for Graylog in the Cloud there is only one way how to change password for root user.

## Change password in Kubernetes Secret

The password for Graylog's root user store in the Kubernetes Secret with the name `graylog-secret` (can be set using
a deployment parameter `graylogSecretName`). And mount inside as an environment variable.

The `logging-operator` watches for changes in the secret named as set in the Secret `graylog-secret`
and has a label `graylog=secret`.

If password in the secret was changed, logging-operator updates `graylog.conf` in the ConfigMap `graylog-service`.

To change password need to:

1. Login in Kubernetes using Kubernetes Dashboard or `kubectl` CLI
2. Navigate to Logging namespace with Graylog
3. Find a Secret with name `graylog-secret`
4. Edit and change password
5. Save it
6. Restart Graylog manually

**Note:** Kubernetes Secret store data base64 encoded. So do not forgot before store a new password encode in base64.
In case of using CLI you can use a command: `echo -n "password" | base64`.
