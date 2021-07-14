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
