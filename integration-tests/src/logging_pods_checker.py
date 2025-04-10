import os
import time

from PlatformLibrary import PlatformLibrary

environ = os.environ
namespace = environ.get("LOGGING_PROJECT")
external_graylog_server = environ.get("EXTERNAL_GRAYLOG_SERVER")
service = 'graylog'
timeout = 300
timeout_before_start = int(environ.get('TIMEOUT_BEFORE_START'))

if __name__ == '__main__':
    print(f'Waiting for {timeout_before_start} seconds')
    time.sleep(timeout_before_start)
    try:
        k8s_lib = PlatformLibrary(managed_by_operator="true")
    except Exception as e:
        print(e)
        exit(1)
    print("Checking Graylog stateful set is ready")
    if external_graylog_server == "true":
        print('Graylog is ready!')
    else:
        timeout_start = time.time()
        while time.time() < timeout_start + timeout:
            try:
                statefulsets = k8s_lib.get_stateful_set_replicas_count(service, namespace)
                ready_statefulsets = k8s_lib.get_stateful_set_ready_replicas_count(service, namespace)
                print(f'[Check status] stateful sets: {statefulsets}, ready stateful sets: {ready_statefulsets}')
            except Exception as e:
                print(e)
                exit(1)
            if statefulsets == ready_statefulsets and statefulsets != 0:
                print("Graylog stateful set is ready")
                break
            time.sleep(10)
        if time.time() >= timeout_start + timeout:
            print(f'Graylog stateful set is not ready at least {timeout} seconds')
            exit(1)

    print("Checking logging agents (fluentd/fluentbit) are ready")
    timeout_start = time.time()
    while time.time() < timeout_start + timeout:
        try:
            daemon_sets = k8s_lib.get_daemon_sets(namespace)
            for daemon_set in daemon_sets:
                name = daemon_set.metadata.name
                numberAvailable = daemon_set.status.numberAvailable
                desiredNumberScheduled = daemon_set.status.desiredNumberScheduled
                print(
                    f'[Check status] {name} daemon sets: {desiredNumberScheduled}, ready daemon sets:'
                    f' {numberAvailable}')
        except Exception as e:
            print(e)
            exit(1)

        if numberAvailable == desiredNumberScheduled and desiredNumberScheduled != 0:
            print("Logging agents are ready")
            exit(0)

        time.sleep(10)
    print(f'Logging agents are not ready at least {timeout} seconds')
    exit(1)
