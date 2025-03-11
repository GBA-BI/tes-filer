## Bio-OS TES Filer

Bio-OS TES Filer is a component of the [Bio-OS task execution service](https://github.com/GBA-BI/bioos-tes) , responsible for handling input and output operations for object storage and DRS data pre- and post-task executor.


#### Deployment
The TES Filer is not a continuously running service within a Kubernetes cluster and does not require a separate deployment. Instead, the image built from this project should be specified as a configuration parameter (filerImage) during the deployment of tes-k8s-agent.

We maintain the TES Filer image at [ghcr.io/GBA-BA/tes-filer](https://github.com/orgs/GBA-BI/packages), but if your Kubernetes computing cluster lacks a reliable image caching mechanism, we strongly recommend copying this image to a registry associated with your computing environment to ensure stability.

## License
This project is licensed under the Apache 2.0 License - see the [LICENSE](LICENSE) file for details.