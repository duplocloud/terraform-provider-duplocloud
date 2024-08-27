resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

resource "duplocloud_ecache_instance" "mycache" {
  tenant_id                  = duplocloud_tenant.myapp.tenant_id
  name                       = "mycache"
  cache_type                 = 0 // Redis
  replicas                   = 1
  size                       = "cache.t2.small"
  enable_cluster_mode        = true  // applicable only for redis
  number_of_shards           = 1     // applicable only for redis
  automatic_failover_enabled = false // enable auto failover
  engine_version             = var.engine_version

  /*
    LogDeliveryConfigurations only supported for engine version above 6.2.0
    LogDeliveryConfigurations:
        list of Log Delivery Configuration.
        LogFormat = text, json
        LogType = slow-log, engine-log
        DestinationType = cloudwatch-logs, kinesis-firehose
    Refer aws: https://docs.aws.amazon.com/AmazonElastiCache/latest/red-ug/CLI_Log.html
  */
  log_delivery_configuration = jsonencode([
    {
      DestinationDetails : {
        "CloudWatchLogsDetails" : {
          "LogGroup" : "/aws/elasticache/redis/UNIQUE_NAME"
        }
      },
      DestinationType : "cloudwatch-logs",
      LogFormat : "TEXT",
      LogType : "engine-log"
    },
    {
      DestinationDetails : {
        "CloudWatchLogsDetails" : {
          "LogGroup" : "/aws/elasticache/redis/UNIQUE_NAME"
        }
      },
      DestinationType : "cloudwatch-logs",
      LogFormat : "TEXT",
      LogType : "slow-log"
    }
  ])


}
