
# Table of Content

* [Table of Content](#table-of-content)
* [Graylog maintenance](#graylog-maintenance)
  * [Logging Backup](#logging-backup)
  * [Eviction Policy Configuration](#eviction-policy-configuration)
  * [Logging Backup Daemon Manual Deployment](#logging-backup-daemon-manual-deployment)
  * [Manual Backup](#manual-backup)
  * [Restore](#restore)
* [Update Fluents' Configmap](#update-fluents-configmap)

# Graylog Maintenance

This section describes the general Graylog maintenance guidelines for maintenance operations such
as hardware upgrade and so on.

## Logging Backup

Logging-backuper provides an ability to backup all logging data (logs, configuration, and so on) and to make restore if required.

Logging-backuper is installed on the same VM as Graylog and operates as a separate Docker container.

Logging-backuper automatically does backups by schedule and puts backups on Logging VM volume.

Typically, logging-backuper is installed in scope of full Logging Service deployment.

The `deploy_logging_backuper: true/false` parameter of `deploy-logging-service` is responsible for it.

The following parameters belong to logging-backuper:

* `backup_storage_volume` - The folder on Logging VM for backups storing

  **Important**: You need to adjust the HDD size on the Logging VM according to the number of backups you want to store.
  Each backup size approximately equals the current log size on HDD. Also, it is advised not to store backups
  on the Logging VM only. The reliability backups should be stored in another storage as well.

* `backup_cron_pattern` - The CRON pattern for backups schedule configuration
* `backup_eviction_policy` - The backup eviction policy configuration. This parameter configures the procedure
  for old backups. For more information, see [Eviction Policy Configuration](#eviction-policy-configuration)
* `granular_eviction_policy` - The granular backup eviction policy configuration. The Logging Service does not support
  granular backups, so this parameter is not used. It exists for future improvements

## Eviction Policy Configuration

Eviction policy is a comma-separated string of policies written as `$start_time/$interval`.
This policy splits all backups older than `$start_time` to numerous time intervals `$interval` time long.
It then deletes all backups in every interval, except the newest one.

For example:

* `1d/7d` policy depicts "take all backups older than one day, split them in groups by 7-days interval,
  and leave only the newest."
* `0/1h` policy depicts "take all backups older than now, split them in groups by 1 hour and leave only the newest."

## Logging Backup Daemon Manual Deployment

The logging-backuper is installed in scope of full Logging Service deployment, but backuper is an optional component
it can be absent on a particular Logging VM.

If required, logging-backuper can be deployed separately using the `deploy-logging-backuper` job on DeployVM.

Navigate to the jobs and specify the values for the following parameters:

* `GRAYLOG_HOST` - The comma-separated list of target Logging VMs
* `GRAYLOG_VOLUME` - The folder on Logging VM that was used for Graylog deployment as the data storage folder
  (`graylog_volume` parameter of `deploy-logging-service` job. `/srv/docker/graylog` by default)
* `BACKUP_STORAGE_PV` - The folder on Logging VM for backups storing
* `USER` - The remote user for the SSH connection
* `DOCKER_IMAGE` - The Docker image of logging-backuper. It can be found in the Docker registry on Deploy VM.
  For example, `local.deployVM:17001/product/prod.platform.logging_logging-backuper:v.0.3.2_20190702-114745`.
* `SSH_KEY` - The private SSH key for access to GRAYLOG_HOST
* `SCHEDULE_CRON_PATTERN` - The CRON pattern for backup schedule configuration
* `BACKUP_EVICTION_POLICY` - The backup eviction policy configuration. This parameter configures the next procedure
  for old backups. For more information, see [Eviction Policy Configuration](#eviction-policy-configuration).
* `GRANULAR_EVICTION_POLICY` - The granular backup eviction policy configuration.
  The Logging Service does not support granular backups, so this parameter is not used. It exists for future improvements.

Click **Run**.

The logging backuper is up and running on LoggingVM.

## Manual Backup

The logging-backuper does the backups by schedule.

You may need to trigger a backup, for instance, before the Logging Service upgrade.

To trigger a backup manually:

1. Navigate to `GRAYLOG_HOST` by SSH
2. Execute the following command:

    ```bash
    curl -X POST localhost:8080/backup
    ```

## Restore

To restore logging from backup:

1. Navigate to `GRAYLOG_HOST` by SSH as root.
2. Navigate to graylog folder in the root home: `cd ~/graylog`
3. Get the list of available backups using the following command:

    ```bash
    curl -X GET localhost:8080/listbackups
    ```

   For example: `20180809T102744` (date and time of backup)

4. Use one of the backup names and run the following script:
  
    ```bash
    ./restore-logging.sh #{backup_name_here}
    ```

5. Wait until the `restore successful` message is displayed

# Update Fluents' Configmap

Since R2024.4 you can manually update Fluents' configuration without restarting pods.
It realised by configmap-reload sidecar in each Fluents' container.
It watches mounted configmap and notifies the target process that the configmap has been changed.
After that configmap-reload sidecar calls REST-api for hot reload Fluents'.
