# `hetzplay` - work in progress

This repository is currently a work in progress. Below is a vision of what it may eventually do.

---

`hetzplay` makes it easy to run a Minecraft server in Hetzner while only **paying for the time it's used**. For servers used only part-time, this can cut the cost of running a powerful server down significantly -- **~87%** for a server used **15 hours per week** - see [the math](#example-cost-savings)

## How it works

Hetzner bills servers even when they're stopped, making it difficult to run a server part-time. `hetzplay` works around this issue by *deleting* your server when not in use, preserving its content with snapshots.

- Wait for someone to ask the Discord bot to start the server
- Create the server from a previously saved snapshot
- Periodically query the server's connected player count 
- When the server is idle for a few minutes:
  - Stop the server
  - Make a snapshot of the stopped server, preserving your data
  - **Delete** the server, stopping its billing

This reduces your Hetzner costs down to:

- Billable hours while the server runs (varies, one hour minimum)
- The server's primary IP address (€0.50/mo as of 9 Aug 23)
- The server's snapshots (€0.011/GB/month as of 9 Aug 23)

### Example cost comparison

All prices/exchange rates are as of **9 Aug 23**. Some assumptions:

- `ccx12` server (2 dedicated vCPUs, 8 GB RAM, €0.032/hr)
- 15 hours per week of usage (67.5 per month)
- server + world storage size: 10 GB
- keeping two snapshots, one as a backup

Without `hetzplay`:

- Server: €21.35/mo
- Primary IPv4 address: €0.50/mo

Total: **€21.85/mo ($23.97/mo)**

With `hetzplay`:

- Server: €2.16/mo
- Primary IPv4 address: €0.50/mo
- Snapshot storage: €0.22/mo

Total: **€2.88/mo ($3.16/mo)**

**Savings: ~87%**

Even if you add in the cost of a separate server to run `hetzplay` itself, 

## Prerequisites

- A Hetzner server that starts a Minecraft server on boot. See [dbrennand/mc-hetzner](https://github.com/dbrennand/mc-hetzner) for an easy way to provision one.
- An API key with write permissions for the project hosting the server.
- An always-on server on which to run this bot. For example, a Rapsberry Pi at home, or a cheap, low-powered Hetzner server works.
  - Note that this server does *not* need to be publicly reachable - it only needs to be able to reach out *to* the internet, so running it at home can be a great option.

## Configuration

Configure the following:

- `hetzner_api_token` (`string`): Hetzner API token. See [Hetzner's docs](https://docs.hetzner.com/cloud/api/getting-started/generating-api-token) to generate one
- `hetzner_server_id` (`int`): ID of the Hetzner server running the Minecraft server, found on the Overview page below the server type
- `discord_bot_token` (`string`): Discord bot API token
- `mc_host` (`string`): hostname or IP address of the Minecraft server
- `mc_port` (`int`, default `25565`): the port of the Minecraft server

You can also configure the following, but you likely won't need to. The defaults are designed with Hetzner's billing model in mind to maximize cost savings:

- `min_runtime` (`int`,  minutes, default `50`): how long the server must run before being stopped
  - Because Hetzner [bills for a minimum of one hour](https://docs.hetzner.com/cloud/billing/faq#how-do-you-bill-your-servers) (and rounds up), the default value ensures a server that's started, but never used, is billed for only a single hour.
- `max_runtime` (`int`, minutes, default `0`): how long before the server is stopped, regardless of player count—use `0` for unlimited maximum runtime.
- `players_query_interval` (`int`, minutes, default `5`): how often to query the player count of the server, starting after `minimum_runtime` minutes
- `min_idle_time` (`int`, minutes, default `5`): how long the server must be idle before it's stopped
- `backup_snap_count` (`int`, default `1`): how many backup snapshots to keep, in addition to the current snapshot
  - It's recommended to keep at least one backup snapshot

## Potential future features

- Support other game servers.
  - It'd need a way to query the active player count.
  - Alternatively, it could require the stopping the server manually via Discord, and periodically remind the user who started it to kindly shut it down.
- Notify when server starts/stops via webhooks.
  - Also notify on a configurable duration - for example "Send a notification if the server has been running for over eight hours."
- Support alternative ways of interacting with `hetzplay` instead of Discord, like webhooks or [an ntfy topic](https://github.com/binwiederhier/ntfy)