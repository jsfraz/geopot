# lucian

L.U.C.I.A.N (Location-based Unauthorized Connection Investigation and Analysis Network) as the (very cool) name suggests is intended for monitoring SSH login attempts and geolocating remote hosts who failed to login and gathering used credentials.

## How it works

The main idea is that you don't use default SSH port to connect to your remote server/VPS. This Docker Compose configuration maps fake server's port 22 to Docker host's public IP and stores unsuccesful login details in Postgres database:

| id    | ip_version | ip_address     | latitude  | longitude  | country_name | country_code | time_zone0 | zip_code | city_name | region_name | is_proxy | continent | continent_code | user | password  | timestamp                     |
|:-----:|:----------:|:--------------:|:---------:|:----------:|:------------:|:------------:|:----------:|:--------:|:---------:|:-----------:|:--------:|:---------:|:--------------:|:----:|:---------:|:-----------------------------:|
| 35086 | 4          | 180.101.88.252 | 31.311365 | 120.617691 | China        | CN           | +08:00     | 215003   | Suzhou    | Jiangsu     | false    | Asia      | AS             | root | nathalie  | 2024-02-28 14:31:56.356 +0100 |
| 35085 | 4          | 180.101.88.252 | 31.311365 | 120.617691 | China        | CN           | +08:00     | 215003   | Suzhou    | Jiangsu     | false    | Asia      | AS             | root | dfvgbh    | 2024-02-28 14:31:56.043 +0100 |
| 35084 | 4          | 180.101.88.252 | 31.311365 | 120.617691 | China        | CN           | +08:00     | 215003   | Suzhou    | Jiangsu     | false    | Asia      | AS             | root | jlo       | 2024-02-28 14:30:55.356 +0100 |
| 35083 | 4          | 180.101.88.252 | 31.311365 | 120.617691 | China        | CN           | +08:00     | 215003   | Suzhou    | Jiangsu     | false    | Asia      | AS             | root | egk       | 2024-02-28 14:30:53.744 +0100 |
| 35082 | 4          | 180.101.88.252 | 31.311365 | 120.617691 | China        | CN           | +08:00     | 215003   | Suzhou    | Jiangsu     | false    | Asia      | AS             | root | 1qaz2wsx@ | 2024-02-28 14:30:53.397 +0100 |

## Deploy

Change database passwords and public IP address in `.env` file and run `sudo ./compose.sh` to build the image and compose the project.

## Future plans

Frontend coming soon! (soon is relative)
