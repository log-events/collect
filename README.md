# collect

Receives logs in syslog format and sends them to elasticsearch.

`collect.yml`:
```yaml
listen: tcp://0.0.0.0:5140
elastic:
  uri: http://localhost:9200
  index-format: logs-2006.01.02
  doc-type: log
  fields:
    '@timestamp': timestamp
    message: message
    host: hostname
    ident: app_name
    pid: proc_id
    ip: structured_data.origin.ip
    awsId: structured_data.origin.enterpriseId
    sequenceId:
      type: int
      field: structured_data.meta.sequenceId
  index:
    settings:
      number_of_shards: 2
      number_of_replicas: 1
    mappings:
      log:
        properties:
          '@timestamp':
            type: date
          sequenceId:
            type: long
```