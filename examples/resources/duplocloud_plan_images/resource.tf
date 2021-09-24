resource "duplocloud_plan_images" "myplan" {
  plan_id = "myplan"

  image {
    name     = "eks-worker"
    image_id = "ami-1234"
    os       = "linux"
    username = "ec2-user"
  }

  image {
    name     = "other-ami"
    image_id = "ami-1234"
    os       = "windows"
    username = "Administrator"
  }
}

# List of Duplo worker AMI(s), as of September 2021:
#
# - us-east-1: ami-0e9ccc73deac92270
# - us-east-2: ami-0ffc14a3fd6c21f00
# - us-west-1: ami-0e862f7151e1fd290
# - us-west-2: ami-0cacaa5d2e248f840
# - ap-northeast-1: ami-0a3574d58f8c1c78f
# - sa-east-1: ami-02ec782a0ea4f0c4c
#