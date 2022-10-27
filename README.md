# R2-Uploader

Simple image uploader for `typora`, using Cloudflare-R2 as storage.

## Usage

1. compile from source by `go build`
2. generate a `config` file and store it in `~/.config/r2_uploader/config.json` (or `%APPDATA%\r2_uploader\config.json` on Windows)
3. configure `typora` to use `r2_uploader` as image uploader

Here is a template for the config

```json5
{
    "account_id": "",   // Account Id  : R2 -> Account ID
    "access_key": "",   // Access Key  : R2 -> Manage R2 API Tokens -> Create API Token
    "secret_key": "",   // Secret Key  : same as above
    "bucket_name": "",  // Bucket Name : Your bucket name
    "public_url": ""    // Public URL  : R2 -> Your Bucket -> Settings -> Public Access -> Public Bucket URL
}
```
