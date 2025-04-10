import io

import paramiko
import yaml
from PlatformLibrary import PlatformLibrary

k8s_lib = PlatformLibrary()


def parse_yaml_file(file_path):
    return yaml.safe_load(open(file_path))


def add_security_context_to_deployment(path_to_file, namespace):
    deployment = parse_yaml_file(path_to_file)
    pods = k8s_lib.get_pods(namespace)
    for pod in pods:
        if 'logging-service-operator' in pod.metadata.name:
            security_context = pod.spec.security_context
            break
    if security_context == None:
        return deployment
    deployment['spec']['template']['spec']['securityContext'] = security_context
    return deployment


def create_ssh_connection(host, user, ssh_key):
    pkey = paramiko.RSAKey.from_private_key(io.StringIO(ssh_key))
    ssh_client = paramiko.SSHClient()
    ssh_client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
    ssh_client.connect(hostname=host, username=user, pkey=pkey)
    return ssh_client


def execute_command_on_vm(ssh_client, command):
    stdin, stdout, stderr = ssh_client.exec_command(command)
    return stdout.channel.recv_exit_status(), stdout.read()
