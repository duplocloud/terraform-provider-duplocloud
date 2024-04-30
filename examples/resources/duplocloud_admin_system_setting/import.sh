# Example: Importing an existing duplo admin setting
#  - *KEY_TYPE* is the type of setting key. Replace any occurrences of '/' with '_SLASH_' if they exist within the KEY_TYPE.
#  - *KEY* is the key name
#
terraform import duplocloud_admin_system_setting.mySetting *KEY_TYPE*/*KEY*
