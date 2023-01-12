pots
====

A tool that fires off Tailscale logs to Pushover.

# Usage

This is intendet to be behind a reverse proxy that provedes HTTPS.

You will need to set three environment variables:

- `POTS_TOKEN`: This is a random string you generate.
- `PUSHOVER_USER`: Your Pushover user token.
- `PUSHOVER_TOKEN`: Your Pushover app token.

## POTS_TOKEN

This variable is used to give Tailscale a unique string that isn't easily guessed.

For example, if you set`POTS_TOKEN=arstarst`, then you would point Tailscale's
webhook at `https://yourdomain/api/arstarst`.
