def get_excluded_tags(environ) -> list:
    external_graylog = environ.get('EXTERNAL_GRAYLOG_SERVER')
    ssh_key = environ.get('SSH_KEY')
    vm_user = environ.get('VM_USER')
    if external_graylog == 'false' or not ssh_key or not vm_user:
        return ['archiving-plugin']
