resource "duplocloud_other_agents" "agents" {
  name = "duplo-agents"

  agent {
    agent_name                       = "CloudWatchAgent_0"
    agent_linux_package_path         = "https://s3.amazonaws.com/amazoncloudwatch-agent/ubuntu/amd64/latest/amazon-cloudwatch-agent.deb"
    linux_agent_install_status_cmd   = "sudo service amazon-cloudwatch-agent status | grep -wc 'running'"
    linux_agent_service_name         = "amazon-cloudwatch-agent"
    linux_agent_uninstall_status_cmd = "OS_FAMILY=$(cat /etc/os-release | grep PRETTY_NAME); if [[ $OS_FAMILY == *'Ubuntu'* ]]; then sudo apt-get purge --yes --force-yes amazon-cloudwatch-agent; else sudo yum remove -y AwsAgent; fi"
    linux_install_cmd                = "OS_FAMILY=$(cat /etc/os-release | grep PRETTY_NAME); if [[ $OS_FAMILY == *'Ubuntu'* ]]; then wget https://s3.amazonaws.com/amazoncloudwatch-agent/ubuntu/amd64/latest/amazon-cloudwatch-agent.deb; sudo dpkg -i -E ./amazon-cloudwatch-agent.deb; sudo wget -O /opt/aws/amazon-cloudwatch-agent/etc/amazon-cloudwatch-agent.json https://cf-templates-3qf987fmmv5g-us-east-2.s3.us-east-2.amazonaws.com/amazon-cloudwatch-agent.json; sudo service amazon-cloudwatch-agent restart; else wget https://s3.amazonaws.com/amazoncloudwatch-agent/amazon_linux/amd64/latest/amazon-cloudwatch-agent.rpm; sudo rpm -U ./amazon-cloudwatch-agent.rpm; sudo wget -O /opt/aws/amazon-cloudwatch-agent/etc/amazon-cloudwatch-agent.json https://cf-templates-3qf987fmmv5g-us-east-2.s3.us-east-2.amazonaws.com/amazon-cloudwatch-agent.json && sudo service amazon-cloudwatch-agent restart; fi"
    windows_agent_service_name       = "amazon-cloudwatch-agent"
  }

}
