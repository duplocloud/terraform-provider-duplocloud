resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

# -----------------------------------------------------------------------
# Example 1: ExtendedS3DestinationConfiguration 
# -----------------------------------------------------------------------

resource "duplocloud_aws_firehose" "extended_s3" {
  tenant_id            = duplocloud_tenant.myapp.tenant_id
  name                 = "test-firehose-extended-s3"
  delivery_stream_type = "DirectPut"

  extended_s3_destination_configuration {
    bucket_arn          = "arn:aws:s3:::duploservices-sep8-forairflow-100000000004"
    role_arn            = "arn:aws:iam::100000000004:role/firehoseRole"
    prefix              = "firehose/extended/"
    error_output_prefix = "firehose/errors/"
    compression_format  = "GZIP"
    s3_backup_mode      = "Disabled"

    buffering_hints {
      size_in_mbs         = 5
      interval_in_seconds = 300
    }

    processing_configuration {
      enabled = true
      processors {
        type = "Lambda"
        parameters {
          parameter_name  = "LambdaArn"
          parameter_value = "arn:aws:lambda:us-west-2:100000000004:function:duploservices-sep8-testlmb:$LATEST"
        }
        parameters {
          parameter_name  = "BufferSizeInMBs"
          parameter_value = "3"
        }
        parameters {
          parameter_name  = "BufferIntervalInSeconds"
          parameter_value = "60"
        }
      }
    }

    cloudwatch_logging_options {
      enabled         = true
      log_group_name  = "/aws/firehose/extended-s3"
      log_stream_name = "S3Delivery"
    }
  }

  # tags = {
  #   env  = "dev"
  #   type = "extended-s3"
  # }
}

# -----------------------------------------------------------------------
# Example 2: KinesisStreamAsSource → ExtendedS3 (active)
# -----------------------------------------------------------------------
resource "duplocloud_aws_firehose" "kinesis_to_s3" {
  tenant_id            = duplocloud_tenant.myapp.tenant_id
  name                 = "test-firehose-kinesis-s3"
  delivery_stream_type = "KinesisStreamAsSource"

  kinesis_source_configuration {
    kinesis_stream_arn = "arn:aws:kinesis:us-west-2:100000000004:stream/tf-test"
    role_arn           = "arn:aws:iam::100000000004:role/firehoseRole"
  }

  extended_s3_destination_configuration {
    bucket_arn         = "arn:aws:s3:::duploservices-sep8-forairflow-100000000004"
    role_arn           = "arn:aws:iam::100000000004:role/firehoseRole"
    prefix             = "firehose/kinesis/"
    compression_format = "GZIP"

    buffering_hints {
      size_in_mbs         = 64
      interval_in_seconds = 350
    }
  }

  #  tags = {
  #    env    = "dev"
  #    source = "kinesis"
  #  }
}


# -----------------------------------------------------------------------
# Example 3: RedshiftDestinationConfiguration
# -----------------------------------------------------------------------
resource "duplocloud_aws_firehose" "redshift" {
  tenant_id            = duplocloud_tenant.myapp.tenant_id
  name                 = "test-firehose-redshift"
  delivery_stream_type = "DirectPut"

  redshift_destination_configuration {
    cluster_jdbcurl = "jdbc:redshift://mycluster.us-west-2.redshift.amazonaws.com:5439/mydb"
    role_arn        = "arn:aws:iam::100000000004:role/firehoseRole"
    username        = "firehose_user"
    password        = "Replace_me"
    retry_duration  = 3000

    copy_command {
      data_table_name    = "firehose_ingest"
      data_table_columns = "col1,col2,col3"
      copy_options       = "delimiter '|'"
    }

    s3_configuration {
      bucket_arn         = "arn:aws:s3:::duploservices-sep8-forairflow-100000000004"
      role_arn           = "arn:aws:iam::100000000004:role/firehoseRole"
      compression_format = "UNCOMPRESSED"

      buffering_hints {
        size_in_mbs         = 5
        interval_in_seconds = 300
      }
    }

    s3_backup_mode = "Enabled"

    s3_backup_configuration {
      bucket_arn         = "arn:aws:s3:::duploservices-sep8-forairflow-100000000004"
      role_arn           = "arn:aws:iam::100000000004:role/firehoseRole"
      compression_format = "GZIP"

      buffering_hints {
        size_in_mbs         = 5
        interval_in_seconds = 350
      }
    }

    processing_configuration {
      enabled = true

      processors {
        type = "Lambda"

        parameters {
          parameter_name  = "LambdaArn"
          parameter_value = "arn:aws:lambda:us-west-2:100000000004:function:myLambda:$LATEST"
        }
      }
    }

    cloudwatch_logging_options {
      enabled         = true
      log_group_name  = "/aws/kinesisfirehose/redshift"
      log_stream_name = "DestinationDelivery"
    }
  }

  # tags = {
  #   env  = "dev"
  #   type = "redshift"
  # }
}


# -----------------------------------------------------------------------
# Example: OpenSearch Destination Configuration
# -----------------------------------------------------------------------
resource "duplocloud_aws_firehose" "opensearch" {
  tenant_id            = duplocloud_tenant.myapp.tenant_id
  name                 = "test-firehose-opensearch"
  delivery_stream_type = "DirectPut"

  opensearch_destination_configuration {
    index_name            = "firehose-logs"
    role_arn              = "arn:aws:iam::100000000004:role/firehoseRole"
    domain_arn            = "arn:aws:es:us-west-2:100000000004:domain/duploservices-sep8-tfcode"
    index_rotation_period = "OneDay"
    retry_duration        = 306
    s3_backup_mode        = "FailedDocumentsOnly"

    buffering_hints {
      size_in_mbs         = 5
      interval_in_seconds = 350
    }

    s3_configuration {
      bucket_arn         = "arn:aws:s3:::duploservices-sep8-forairflow-100000000004"
      role_arn           = "arn:aws:iam::100000000004:role/firehoseRole"
      compression_format = "UNCOMPRESSED"

      buffering_hints {
        size_in_mbs         = 5
        interval_in_seconds = 300
      }
    }
    vpc_configuration {
      role_arn = "arn:aws:iam::100000000004:role/firehoseRole"
      security_group_ids = [
        "sg-060ea8b713daf0bf2",
      ]
      subnet_ids = [
        "subnet-065ab3e894092dd1c",
        "subnet-09252308e1a093bda",
      ]
    }
    cloudwatch_logging_options {
      enabled         = true
      log_group_name  = "/aws/firehose/opensearch"
      log_stream_name = "DestinationDelivery"
    }
  }
}

# -----------------------------------------------------------------------
# Example: OpenSearch Serverless Destination Configuration
# -----------------------------------------------------------------------
resource "duplocloud_aws_firehose" "opensearch_serverless" {
  tenant_id            = duplocloud_tenant.myapp.tenant_id
  name                 = "test-firehose-oss"
  delivery_stream_type = "DirectPut"

  opensearch_serverless_destination_configuration {
    collection_endpoint = "https://tuvxe7y4ampqqf0xjtse.us-west-2.aoss.amazonaws.com"
    index_name          = "firehose-logs"
    role_arn            = "arn:aws:iam::100000000004:role/firehoseRole"
    retry_duration      = 300
    s3_backup_mode      = "FailedDocumentsOnly"

    buffering_hints {
      size_in_mbs         = 5
      interval_in_seconds = 300
    }

    s3_configuration {
      bucket_arn = "arn:aws:s3:::duploservices-sep8-forairflow-100000000004"
      role_arn   = "arn:aws:iam::100000000004:role/firehoseRole"

      buffering_hints {
        size_in_mbs         = 5
        interval_in_seconds = 350
      }
    }

    vpc_configuration {
      role_arn = "arn:aws:iam::100000000004:role/firehoseRole"
      security_group_ids = [
        "sg-060ea8b713daf0bf2",
      ]
      subnet_ids = [
        "subnet-065ab3e894092dd1c",
        "subnet-09252308e1a093bda",
      ]
    }

    cloudwatch_logging_options {
      enabled         = true
      log_group_name  = "/aws/firehose/opensearch-serverless"
      log_stream_name = "DestinationDelivery"
    }
  }
}


# -----------------------------------------------------------------------
# Example: Elasticsearch Destination Configuration
# -----------------------------------------------------------------------
resource "duplocloud_aws_firehose" "elasticsearch" {
  tenant_id            = duplocloud_tenant.myapp.tenant_id
  name                 = "test-firehose-es"
  delivery_stream_type = "DirectPut"

  elasticsearch_destination_configuration {
    index_name            = "firehose-logs"
    role_arn              = "arn:aws:iam::100000000004:role/firehoseRole"
    domain_arn            = "arn:aws:es:us-west-2:100000000004:domain/duploservice-sep8-es"
    index_rotation_period = "OneDay"
    retry_duration        = 300
    s3_backup_mode        = "FailedDocumentsOnly"

    buffering_hints {
      size_in_mbs         = 5
      interval_in_seconds = 300
    }

    s3_configuration {
      bucket_arn = "arn:aws:s3:::duploservices-sep8-forairflow-100000000004"
      role_arn   = "arn:aws:iam::100000000004:role/firehoseRole"

      buffering_hints {
        size_in_mbs         = 5
        interval_in_seconds = 300
      }
    }

    cloudwatch_logging_options {
      enabled         = true
      log_group_name  = "/aws/firehose/elasticsearch"
      log_stream_name = "DestinationDelivery"
    }
    vpc_configuration {
      role_arn = "arn:aws:iam::100000000004:role/firehoseRole"
      security_group_ids = [
        "sg-060ea8b713daf0bf2",
      ]
      subnet_ids = [
        "subnet-065ab3e894092dd1c",
        "subnet-09252308e1a093bda",
      ]
    }

  }
}


# -----------------------------------------------------------------------
# Example: MSK Source Configuration (with ExtendedS3 destination)
# -----------------------------------------------------------------------
resource "duplocloud_aws_firehose" "msk_source" {
  tenant_id            = duplocloud_tenant.myapp.tenant_id
  name                 = "test-firehose-msk"
  delivery_stream_type = "MSKAsSource"

  msk_source_configuration {
    msk_cluster_arn     = "arn:aws:kafka:us-west-2:100000000004:cluster/duploservices-sep8-msk/bedc1eeb-cd0f-4379-bbb4-8fb3cfd816d5-s3"
    topic_name          = "my-kafka-topic"
    read_from_timestamp = "2026-03-31T00:00:00Z"

    authentication_configuration {
      connectivity = "PRIVATE"
      role_arn     = "arn:aws:iam::100000000004:role/firehoseRole"
    }
  }

  extended_s3_destination_configuration {
    bucket_arn = "arn:aws:s3:::duploservices-sep8-forairflow-100000000004"
    role_arn   = "arn:aws:iam::100000000004:role/firehoseRole"

    buffering_hints {
      size_in_mbs         = 5
      interval_in_seconds = 350
    }
  }
}
