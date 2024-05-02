# Triblerr \*arr apps shim

Provides a qBittorrent API shim for [Tribler](https://github.com/tribler/tribler).

# What's that for?

If you're using *arr apps and you want to use Tribler as a download client, you can use this shim to allow *arr apps to communicate with Tribler through configuring it as a qBittorrent client.

# Configuration

Tribler shim includes the following configuration options:

- TRIBLER_ARR_SHIM_SCHEME="http"
- TRIBLER_ARR_SHIM_ADDR="localhost"
- TRIBLER_ARR_SHIM_PORT="8091"
- TRIBLER_API_ENDPOINT="localhost:20100"
- TRIBLER_API_KEY=""
- TRIBLER_DOWNLOAD_DIR="/downloads"
- DEFAULT_CATEGORY=""
- TLS_SKIP_VERIFY="false"

# Run as a Docker container

1. Deploy Tribler
Instructions on deploying Tribler is beyond the scope of this document.
Please refer to official Tribler instructions - [Tribler](https://github.com/tribler/tribler)

Ensure the shim can access the endpoint that Tribler API is running on, and note down the API key.

1. Copy .env.example to .env

```bash
cp .env.example .env

```

1. Modify .env variables as needed
Fill in Tribler API details and preferred downloads dir.
DEFAULT_CATEGORY can be arbitrary, it only used to return the save path for *arr apps.


1. Run Docker image
```bash
docker --env-from .env run github.com/sashkachan/tribler-arr-shim:main
```

# Caveats
## Categories
*arr apps distinguish their downloads from others by fetching the list with a category selector that will be applied to all API calls.
Tribler does not implement a concept of categories but it can add tags to downloads.
This means that regardless of the configured category in an *arr app, all downloads will be returned on list requests. 

# TODO
1. Support multiple categories. 


## License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details.
