# do-image-uploader

An easy to use DigitalOcean custom image uploader.



## Installation

```shell
go get github.com/htr/do-image-uploader
```


## Usage

Build a VM image in any of the [supported formats](https://www.digitalocean.com/docs/images/custom-images/). You can get some ideas [here](https://github.com/htr/vm-image-builder).

Upload the image:
```
$ do-image-uploader --api-token=... --image-file=image.qcow2.gz --region=fra1 --name=my-image --wait-until-available
```

Currently, the uploader assumes there is public IP connectivity. In the future, image upload through DO Spaces might be implemented.


